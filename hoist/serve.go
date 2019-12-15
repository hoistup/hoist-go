package hoist

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/JosiahWitt/erk"
	"github.com/JosiahWitt/erk/erg"
	"github.com/hoistup/hoist-go/erks"
	"github.com/hoistup/hoist-go/strand"
	"github.com/hoistup/hoist-go/wire"
)

type (
	ErkHoistInit  struct{ erks.Default }
	ErkBadRequest struct{ erks.Default }
)

var (
	ErrInitializing      = erk.New(ErkHoistInit{}, "could not initialize")
	ErrPortMissing       = erk.New(ErkHoistInit{}, "PORT environment variable not set (this should be set automatically by Hoist)")
	ErrJSONParamsInvalid = erk.New(ErkBadRequest{}, "function params are invalid JSON")
)

// Serve the Hoisted application.
//
// Serve blocks with a call to http.Server{}.ListenAndServe().
func (s *Service) Serve() error {
	if errs := s.Errors(); len(errs) > 0 {
		return erg.NewAs(ErrInitializing, errs...)
	}

	port := os.Getenv("PORT")
	if port == "" {
		return ErrPortMissing
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/_/v1/fn", s.handler)

	server := http.Server{
		Addr:           "localhost:" + port,
		Handler:        mux,
		MaxHeaderBytes: 250,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
	}

	fmt.Printf("Serving at http://localhost:%s\n", port)
	return server.ListenAndServe()
}

func (s *Service) handler(w http.ResponseWriter, r *http.Request) {
	details, err := s.handleHTTPEvent(w, r)

	// Handle error, if present
	if err != nil {
		errDetails := &strand.ResponseDetails{IsError: true}
		if details != nil {
			errDetails.RequestID = details.RequestID
		}

		// Export the error
		params, isInternalError := s.exportEventError(err)
		errDetails.IsInternalError = isInternalError

		// Encode the error
		bytes, err := wire.Encode(errDetails, params)
		if err != nil {
			errDetails := `{"err":true,"ierr":true}`
			errParams := `{"kind":"wire_encoding_error","message":"unable to encode error to wire format"}`
			w.Write([]byte(`1,24,80:` + errDetails + errParams))
			return
		}

		// Write the error
		w.Write(bytes)
	}
}

func (s *Service) handleHTTPEvent(w http.ResponseWriter, r *http.Request) (*strand.RequestDetails, error) {
	decoded, err := wire.NewDecoder(r.Body).Decode()
	if err != nil {
		return nil, err
	}

	details := strand.RequestDetails{}
	if err := json.Unmarshal(decoded.RawDetails, &details); err != nil {
		return nil, erk.WrapAs(ErrJSONParamsInvalid, err)
	}

	result, err := s.Call(details.FunctionName, decoded.RawParams)
	if err != nil {
		return &details, err
	}

	bytes, err := wire.Encode(&strand.ResponseDetails{RequestID: details.RequestID}, result)
	if err != nil {
		return &details, err
	}

	_, err = w.Write(bytes)
	return &details, err
}

func (s *Service) exportEventError(err error) (interface{}, bool) {
	// Check if the error denotes the function call failed
	if errors.Is(err, ErrFunctionCallFailed) {
		// Attempt to unwrap the error
		wrappedErr := errors.Unwrap(err)
		if wrappedErr == nil {
			return erk.Export(err), true
		}

		// Check if an "Export" method exists with no input params and one return value
		exportMethod := reflect.ValueOf(wrappedErr).MethodByName("Export")
		if exportMethod.IsValid() && exportMethod.Type().NumOut() == 1 && exportMethod.Type().NumIn() == 0 {
			// Call the "Export" method
			ret := exportMethod.Call([]reflect.Value{})
			return ret[0].Interface(), false
		}

		// Fallback to calling the Error method
		return wrappedErr.Error(), false
	}

	return erk.Export(err), true
}

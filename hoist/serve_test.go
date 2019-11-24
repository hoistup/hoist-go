package hoist_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/JosiahWitt/erk"
	"github.com/hoistup/hoist-go/hoist"
	"github.com/hoistup/hoist-go/strand"
	"github.com/hoistup/hoist-go/wire"
	"github.com/matryer/is"
	"github.com/phayes/freeport"
)

const reqID = "my-id"

var (
	errRespDetailsNotInternal = strand.ResponseDetails{
		RequestID:       reqID,
		IsError:         true,
		IsInternalError: false,
	}
	errRespDetailsInternal = strand.ResponseDetails{
		IsError:         true,
		IsInternalError: true,
	}
	errRespDetailsInternalReqID = strand.ResponseDetails{
		RequestID:       reqID,
		IsError:         true,
		IsInternalError: true,
	}
)

type TestContext struct{}
type TestParams struct{ Message string }

type ExportedError struct{}

func (ExportedError) Error() string {
	return "ExportedError.Error()"
}

func (ExportedError) Export() string {
	return "ExportedError.Export()"
}

type ErkError erk.DefaultKind

var ErrErkError = erk.New(ErkError{}, "my erk error")

func TestServe(t *testing.T) {
	t.Run("when errors exist before starting the server", setPORT(func(t *testing.T) {
		is := is.New(t)

		s := hoist.NewService("abc")
		s.RegisterAs("my-func", nil)
		is.True(errors.Is(s.Serve(), hoist.ErrInitializing))
	}))

	t.Run("when PORT not set", func(t *testing.T) {
		is := is.New(t)

		originalPORT := os.Getenv("PORT")
		os.Unsetenv("PORT")

		s := hoist.NewService("abc")
		is.True(errors.Is(s.Serve(), hoist.ErrPortMissing))

		os.Setenv("PORT", originalPORT)
	})

	t.Run("with running service", setPORT(func(t *testing.T) {
		startService()

		t.Run("with valid service", func(t *testing.T) {
			is := is.New(t)

			reqDetails := strand.RequestDetails{
				RequestID:    reqID,
				ServiceName:  "abc",
				FunctionName: "echo",
			}
			const message = "testing 123"
			reqParams := TestParams{Message: message}

			respDetails, respParams, _, err := makeRequest(&reqDetails, &reqParams)
			is.NoErr(err)

			is.Equal(respDetails, &strand.ResponseDetails{
				RequestID: reqID,
			})

			is.Equal(respParams, &TestParams{
				Message: "echo: " + message,
			})
		})

		t.Run("with nil body", func(t *testing.T) {
			is := is.New(t)

			strands, err := makeRawRequest(nil)
			is.NoErr(err)
			respDetails, _, err := parseRequest(strands)
			is.NoErr(err)

			is.Equal(respDetails, &errRespDetailsInternal)
			errEqual(is, strands, wire.ErrUnableToReadInfoHeader)
		})

		t.Run("with invalid JSON request details", func(t *testing.T) {
			is := is.New(t)

			strands, err := makeRawRequest([]byte("1,8,4:not-jsonnull"))
			is.NoErr(err)
			respDetails, _, err := parseRequest(strands)
			is.NoErr(err)

			is.Equal(respDetails, &errRespDetailsInternal)
			errEqual(is, strands, hoist.ErrJSONParamsInvalid)
		})

		t.Run("with invalid function name", func(t *testing.T) {
			is := is.New(t)

			reqDetails := strand.RequestDetails{
				RequestID:    reqID,
				ServiceName:  "abc",
				FunctionName: "dne",
			}

			respDetails, _, strands, err := makeRequest(&reqDetails, nil)
			is.NoErr(err)

			is.Equal(respDetails, &errRespDetailsInternalReqID)
			errEqual(is, strands, hoist.ErrFunctionNotFound, "service 'abc' does not have function 'dne'")
		})

		t.Run("with function result that cannot be encoded", func(t *testing.T) {
			is := is.New(t)

			reqDetails := strand.RequestDetails{
				RequestID:    reqID,
				ServiceName:  "abc",
				FunctionName: "bad-fn",
			}

			respDetails, _, strands, err := makeRequest(&reqDetails, nil)
			is.NoErr(err)

			is.Equal(respDetails, &errRespDetailsInternalReqID)
			errEqual(is, strands, wire.ErrUnableToEncodeParams)
		})

		t.Run("with function that returns errors.New() error", func(t *testing.T) {
			is := is.New(t)

			reqDetails := strand.RequestDetails{
				RequestID:    reqID,
				ServiceName:  "abc",
				FunctionName: "errors.New()",
			}

			respDetails, _, strands, err := makeRequest(&reqDetails, nil)
			is.NoErr(err)

			is.Equal(respDetails, &errRespDetailsNotInternal)

			errParams, err := json.Marshal("an error")
			is.NoErr(err)
			is.Equal(strands.RawParams, errParams)
		})

		t.Run("with function that returns error with Export method", func(t *testing.T) {
			is := is.New(t)

			reqDetails := strand.RequestDetails{
				RequestID:    reqID,
				ServiceName:  "abc",
				FunctionName: "exported-error",
			}

			respDetails, _, strands, err := makeRequest(&reqDetails, nil)
			is.NoErr(err)

			is.Equal(respDetails, &errRespDetailsNotInternal)

			errParams, err := json.Marshal("ExportedError.Export()")
			is.NoErr(err)
			is.Equal(strands.RawParams, errParams)
		})

		t.Run("with function that returns erk error", func(t *testing.T) {
			is := is.New(t)

			reqDetails := strand.RequestDetails{
				RequestID:    reqID,
				ServiceName:  "abc",
				FunctionName: "erk-error",
			}

			respDetails, _, strands, err := makeRequest(&reqDetails, nil)
			is.NoErr(err)

			is.Equal(respDetails, &errRespDetailsNotInternal)
			errEqual(is, strands, ErrErkError)
		})

		t.Run("with function that returns unmarshalable error", func(t *testing.T) {
			is := is.New(t)

			reqDetails := strand.RequestDetails{
				RequestID:    reqID,
				ServiceName:  "abc",
				FunctionName: "unmarshalable-error",
			}

			respDetails, _, strands, err := makeRequest(&reqDetails, nil)
			is.NoErr(err)

			is.Equal(respDetails, &errRespDetailsInternal)
			is.Equal(string(strands.RawParams), `{"kind":"wire_encoding_error","message":"unable to encode error to wire format"}`)
		})
	}))
}

func setPORT(fn func(t *testing.T)) func(t *testing.T) {
	return func(t *testing.T) {
		port, err := freeport.GetFreePort()
		if err != nil {
			t.Errorf("while getting free port, expected no err, got: %v", err)
		}

		originalPORT := os.Getenv("PORT")
		os.Setenv("PORT", strconv.Itoa(port))
		fn(t)
		os.Setenv("PORT", originalPORT)
	}
}

func startService() {
	s := hoist.NewService("abc")

	s.RegisterAs("echo", func(ctx *TestContext, params *TestParams) (*TestParams, error) {
		params.Message = "echo: " + params.Message
		return params, nil
	})

	s.RegisterAs("bad-fn", func(ctx *TestContext, params *TestParams) (chan int, error) {
		return make(chan int), nil
	})

	s.RegisterAs("errors.New()", func(ctx *TestContext, params *TestParams) (chan int, error) {
		return nil, errors.New("an error")
	})

	s.RegisterAs("exported-error", func(ctx *TestContext, params *TestParams) (chan int, error) {
		return nil, ExportedError{}
	})

	s.RegisterAs("erk-error", func(ctx *TestContext, params *TestParams) (chan int, error) {
		return nil, ErrErkError
	})

	s.RegisterAs("unmarshalable-error", func(ctx *TestContext, params *TestParams) (chan int, error) {
		return nil, erk.WithParam(ErrErkError, "unmarshalable", make(chan int))
	})

	// Start the service in a goroutine
	// TODO: Shutdown the service
	go func() {
		err := s.Serve()
		if err != nil {
			log.Fatalf("server shutdown with: %v\n", err)
		}
	}()

	// Wait until the server starts
	deadline := time.Now().Add(30 * time.Second)
	for {
		_, err := http.Get(fmt.Sprintf("http://localhost:%s", os.Getenv("PORT")))
		if err == nil {
			break
		}

		if time.Now().After(deadline) {
			log.Fatalln("server did not start for hoist/serve_test.go")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func makeRequest(reqDetails *strand.RequestDetails, reqParams *TestParams) (*strand.ResponseDetails, *TestParams, *wire.DecodeResult, error) {
	// Encode the request
	body, err := wire.Encode(&reqDetails, &reqParams)
	if err != nil {
		return nil, nil, nil, err
	}

	// Make the request
	strands, err := makeRawRequest(body)
	if err != nil {
		return nil, nil, nil, err
	}

	respDetails, respParams, err := parseRequest(strands)
	return respDetails, respParams, strands, err
}

func makeRawRequest(body []byte) (*wire.DecodeResult, error) {
	// Setup the request
	client := http.Client{}
	url := fmt.Sprintf("http://localhost:%s/_/v1/fn", os.Getenv("PORT"))
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Do the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Decode the result
	return wire.NewDecoder(resp.Body).Decode()
}

func parseRequest(strands *wire.DecodeResult) (*strand.ResponseDetails, *TestParams, error) {
	// Unmarshal the details
	var respDetails strand.ResponseDetails
	err := json.Unmarshal(strands.RawDetails, &respDetails)
	if err != nil {
		return nil, nil, err
	}

	// If response is an error, or the params are nil, return only the details
	if respDetails.IsError || strands.RawParams == nil {
		return &respDetails, nil, nil
	}

	// Unmarshal the params
	var respParams TestParams
	err = json.Unmarshal(strands.RawParams, &respParams)
	if err != nil {
		return nil, nil, err
	}

	return &respDetails, &respParams, nil
}

func errEqual(is *is.I, strands *wire.DecodeResult, expected error, message ...string) {
	var respErr erk.ExportedError
	err := json.Unmarshal(strands.RawParams, &respErr)
	is.NoErr(err)

	// Export the expected error
	expectedExported := erk.Export(expected)
	is.Equal(respErr.Kind, expectedExported.ErrorKind())

	// If a message is present, check equality with the message,
	// otherwise check against the error message
	if len(message) > 0 {
		is.Equal(respErr.Message, message[0])
	} else {
		is.Equal(respErr.Message, expectedExported.ErrorMessage())
	}
}

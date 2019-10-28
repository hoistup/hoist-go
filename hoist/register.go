package hoist

import (
	"encoding/json"
	"reflect"

	"github.com/JosiahWitt/erk"
)

type (
	ErkInvalidFunction erk.DefaultKind
	ErkFunctionCall    erk.DefaultKind
)

var (
	ErrInvalidParameterNumber    = erk.New(ErkInvalidFunction{}, "function must have exactly two parameters: (context, params), got: {{.numParams}}")
	ErrInvalidReturnNumber       = erk.New(ErkInvalidFunction{}, "function must have exactly two return values: (data, error), got: {{.numReturns}}")
	ErrInvalidReturnMissingError = erk.New(ErkInvalidFunction{}, "second return value must be an error")
	ErrNotFunction               = erk.New(ErkInvalidFunction{}, "a function was not provided")
	ErrInvalidFunction           = erk.New(ErkInvalidFunction{}, "service '{{.serviceName}}' could not register function '{{.fnName}}': {{.err}}")

	ErrFunctionCallJSONUnmarshal = erk.New(ErkFunctionCall{}, "could not unmarshal JSON '{{.originalJSON}}'")
)

// errorType allows us to check that the second parameter is an error.
var errorType = reflect.TypeOf(new(error)).Elem()

// RegisterAs allows you to register a function as the provided name.
//
// fn must be a function with the following signature:
//  func myFunction(ctx myContextType, params myParamType) (myDataType, error)
func (s *Service) RegisterAs(fnName string, fn interface{}) {
	// Wrap the function
	wrappedFn, err := s.funcWrapper(fn)

	// Acquire the lock
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for an error
	if err != nil {
		wrappedErr := erk.WithParams(erk.WrapAs(ErrInvalidFunction, err), erk.Params{"fnName": fnName, "serviceName": s.name})
		s.errors = append(s.errors, wrappedErr)
		return
	}

	// Add the function
	s.funcs[fnName] = wrappedFn
}

func (s *Service) funcWrapper(fn interface{}) (rawFunc, error) {
	// Get the function type
	fnType := reflect.TypeOf(fn)
	if fnType == nil || fnType.Kind() != reflect.Func {
		return nil, ErrNotFunction
	}

	// Check the parameter and return counts
	if fnType.NumIn() != 2 {
		return nil, erk.WithParam(ErrInvalidParameterNumber, "numParams", fnType.NumIn())
	}
	if fnType.NumOut() != 2 {
		return nil, erk.WithParam(ErrInvalidReturnNumber, "numReturns", fnType.NumOut())
	}

	// Check the second return value
	if !fnType.Out(1).Implements(errorType) {
		return nil, ErrInvalidReturnMissingError
	}

	// Create the function
	wrappedFn := func(rawParams []byte) (interface{}, error) {
		// Create the context and params
		ctx := reflect.New(fnType.In(0))
		params := reflect.New(fnType.In(1))

		// Fill the params with the JSON
		if err := json.Unmarshal(rawParams, params.Interface()); err != nil {
			return nil, erk.WithParam(erk.WrapAs(ErrFunctionCallJSONUnmarshal, err), "originalJSON", string(rawParams))
		}

		// TODO: Fill ctx with hooks

		// Call the function
		rets := reflect.ValueOf(fn).Call([]reflect.Value{ctx.Elem(), params.Elem()})

		// Check for an error
		if err := rets[1].Interface(); err != nil {
			err, ok := err.(error)
			if !ok {
				return nil, ErrInvalidReturnMissingError
			}
			if err != nil {
				return nil, err
			}
		}

		// Return the data
		return rets[0].Interface(), nil
	}

	return wrappedFn, nil
}

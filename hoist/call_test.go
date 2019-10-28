package hoist_test

import (
	"errors"
	"testing"

	"github.com/hoistup/hoist-go/hoist"
	"github.com/matryer/is"
)

func TestCall(t *testing.T) {
	const serviceName = "myService"
	var myErr = errors.New("it's an error")

	type MyReturn struct {
		MyVal string
	}

	table := []struct {
		Name          string
		Before        func(*hoist.Service)
		FnName        string
		RawParams     string
		ExpectedData  interface{}
		ExpectedError error
	}{
		{
			Name: "valid call",
			Before: func(s *hoist.Service) {
				type MyParams struct {
					Abc string
				}

				s.RegisterAs("myFunc", func(ctx *MyCtx, params *MyParams) (*MyReturn, error) {
					return &MyReturn{
						MyVal: params.Abc + "!",
					}, nil
				})
			},
			FnName:    "myFunc",
			RawParams: `{"abc": "hi"}`,
			ExpectedData: &MyReturn{
				MyVal: "hi!",
			},
			ExpectedError: nil,
		},
		{
			Name: "function not found",
			Before: func(s *hoist.Service) {
				s.RegisterAs("myFunc", func(*MyCtx, *MyParams) (*MyData, error) {
					return nil, nil
				})
			},
			FnName:        "myFunc2",
			RawParams:     `{}`,
			ExpectedData:  nil,
			ExpectedError: hoist.ErrFunctionNotFound,
		},
		{
			Name: "JSON unmarshal error",
			Before: func(s *hoist.Service) {
				type MyParams struct {
					Abc int
				}

				s.RegisterAs("myFunc", func(*MyCtx, *MyParams) (*MyData, error) {
					return nil, nil
				})
			},
			FnName:        "myFunc",
			RawParams:     `{"abc": "hi"}`,
			ExpectedData:  nil,
			ExpectedError: hoist.ErrFunctionCallJSONUnmarshal,
		},
		{
			Name: "function error",
			Before: func(s *hoist.Service) {
				type MyParams struct {
					Abc string
				}

				s.RegisterAs("myFunc", func(*MyCtx, *MyParams) (*MyData, error) {
					return nil, myErr
				})
			},
			FnName:        "myFunc",
			RawParams:     `{"abc": "hi"}`,
			ExpectedData:  nil,
			ExpectedError: myErr,
		},
	}

	for _, entry := range table {
		t.Run(entry.Name, func(t *testing.T) {
			is := is.New(t)

			s := hoist.NewService(serviceName)
			entry.Before(s)
			data, err := s.Call(entry.FnName, []byte(entry.RawParams))
			is.True(errors.Is(err, entry.ExpectedError))
			is.Equal(data, entry.ExpectedData)
		})
	}
}

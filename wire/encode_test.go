package wire_test

import (
	"errors"
	"testing"

	"github.com/hoistup/hoist-go/wire"
	"github.com/matryer/is"
)

func TestEncode(t *testing.T) {
	type MyDetails struct {
		ServiceName string `json:"svc"`
		FuncName    string `json:"fn"`
	}

	table := []struct {
		Name             string
		Details          interface{}
		Params           interface{}
		ExpectedEncoding string
		ExpectedError    error
	}{
		{
			Name:          "with nil details and nil params",
			Details:       nil,
			Params:        nil,
			ExpectedError: wire.ErrNilDetails,
		},
		{
			Name: "with present details and nil params",
			Details: MyDetails{
				ServiceName: "myService",
				FuncName:    "myFunc",
			},
			Params:           nil,
			ExpectedEncoding: `1,33,4:{"svc":"myService","fn":"myFunc"}null`,
		},
		{
			Name: "with present details and present string params",
			Details: MyDetails{
				ServiceName: "myService2",
				FuncName:    "myFunc2",
			},
			Params:           "a param",
			ExpectedEncoding: `1,35,9:{"svc":"myService2","fn":"myFunc2"}"a param"`,
		},
		{
			Name: "with present details and present map params",
			Details: MyDetails{
				ServiceName: "myService2",
				FuncName:    "myFunc2",
			},
			Params: map[string]interface{}{
				"Param1": "abc",
				"Param2": "xyz",
			},
			ExpectedEncoding: `1,35,31:{"svc":"myService2","fn":"myFunc2"}{"Param1":"abc","Param2":"xyz"}`,
		},
		{
			Name: "with present details and present struct params",
			Details: MyDetails{
				ServiceName: "myService",
				FuncName:    "myFunc2",
			},
			Params: struct {
				Message string `json:"msg"`
			}{
				Message: "hello",
			},
			ExpectedEncoding: `1,34,15:{"svc":"myService","fn":"myFunc2"}{"msg":"hello"}`,
		},
		{
			Name:          "with invalid details",
			Details:       make(chan struct{}),
			Params:        "a string",
			ExpectedError: wire.ErrUnableToEncodeDetails,
		},
		{
			Name: "with invalid params",
			Details: MyDetails{
				ServiceName: "myService",
				FuncName:    "myFunc2",
			},
			Params:        make(chan struct{}),
			ExpectedError: wire.ErrUnableToEncodeParams,
		},
	}

	for _, entry := range table {
		t.Run(entry.Name, func(t *testing.T) {
			is := is.New(t)

			encoding, err := wire.Encode(entry.Details, entry.Params)
			is.True(errors.Is(err, entry.ExpectedError))
			is.Equal(string(encoding), entry.ExpectedEncoding)
		})
	}
}

package hoist_test

import (
	"errors"
	"testing"

	"github.com/hoistup/hoist-go/hoist"
	"github.com/matryer/is"
)

func TestRegisterAs(t *testing.T) {
	const serviceName = "myService"

	table := []struct {
		Name  string
		Check func(is *is.I, s *hoist.Service)
	}{
		{
			Name: "valid functions",
			Check: func(is *is.I, s *hoist.Service) {
				name := "myFn"
				name2 := "myFn2"
				s.RegisterAs(name, validNoopFn)
				s.RegisterAs(name2, validNoopFn)

				is.Equal(len(s.Errors()), 0)

				expected := &hoist.ExportedService{
					Name: serviceName,
					Functions: map[string]*hoist.ExportedFunction{
						name: {
							Name: name,
						},
						name2: {
							Name: name2,
						},
					},
				}
				is.Equal(s.Export(), expected)
			},
		},
		{
			Name: "invalid function: nil function",
			Check: func(is *is.I, s *hoist.Service) {
				name := "myFn"
				name2 := "myFn2"
				s.RegisterAs(name, nil)
				s.RegisterAs(name2, validNoopFn)

				is.Equal(len(s.Errors()), 1)
				is.True(errors.Is(s.Errors()[0], hoist.ErrNotFunction))

				expected := &hoist.ExportedService{
					Name: serviceName,
					Functions: map[string]*hoist.ExportedFunction{
						name2: {
							Name: name2,
						},
					},
				}
				is.Equal(s.Export(), expected)
			},
		},
		{
			Name: "invalid function: not a function",
			Check: func(is *is.I, s *hoist.Service) {
				name := "myFn"
				name2 := "myFn2"
				s.RegisterAs(name, "a string")
				s.RegisterAs(name2, validNoopFn)

				is.Equal(len(s.Errors()), 1)
				is.True(errors.Is(s.Errors()[0], hoist.ErrNotFunction))

				expected := &hoist.ExportedService{
					Name: serviceName,
					Functions: map[string]*hoist.ExportedFunction{
						name2: {
							Name: name2,
						},
					},
				}
				is.Equal(s.Export(), expected)
			},
		},
		{
			Name: "invalid function: too few params",
			Check: func(is *is.I, s *hoist.Service) {
				name := "myFn"
				name2 := "myFn2"

				badFn := func(*MyParams) (*MyData, error) {
					return nil, nil
				}
				s.RegisterAs(name, badFn)
				s.RegisterAs(name2, validNoopFn)

				is.Equal(len(s.Errors()), 1)
				is.True(errors.Is(s.Errors()[0], hoist.ErrInvalidParameterNumber))

				expected := &hoist.ExportedService{
					Name: serviceName,
					Functions: map[string]*hoist.ExportedFunction{
						name2: {
							Name: name2,
						},
					},
				}
				is.Equal(s.Export(), expected)
			},
		},
		{
			Name: "invalid function: too many params",
			Check: func(is *is.I, s *hoist.Service) {
				name := "myFn"
				name2 := "myFn2"

				badFn := func(*MyParams, *MyParams, *MyParams) (*MyData, error) {
					return nil, nil
				}
				s.RegisterAs(name, badFn)
				s.RegisterAs(name2, validNoopFn)

				is.Equal(len(s.Errors()), 1)
				is.True(errors.Is(s.Errors()[0], hoist.ErrInvalidParameterNumber))

				expected := &hoist.ExportedService{
					Name: serviceName,
					Functions: map[string]*hoist.ExportedFunction{
						name2: {
							Name: name2,
						},
					},
				}
				is.Equal(s.Export(), expected)
			},
		},
		{
			Name: "invalid function: too few return values",
			Check: func(is *is.I, s *hoist.Service) {
				name := "myFn"
				name2 := "myFn2"

				badFn := func(*MyCtx, *MyParams) error {
					return nil
				}
				s.RegisterAs(name, badFn)
				s.RegisterAs(name2, validNoopFn)

				is.Equal(len(s.Errors()), 1)
				is.True(errors.Is(s.Errors()[0], hoist.ErrInvalidReturnNumber))

				expected := &hoist.ExportedService{
					Name: serviceName,
					Functions: map[string]*hoist.ExportedFunction{
						name2: {
							Name: name2,
						},
					},
				}
				is.Equal(s.Export(), expected)
			},
		},
		{
			Name: "invalid function: too many return values",
			Check: func(is *is.I, s *hoist.Service) {
				name := "myFn"
				name2 := "myFn2"

				badFn := func(*MyCtx, *MyParams) (*MyData, error, *MyData) {
					return nil, nil, nil
				}
				s.RegisterAs(name, badFn)
				s.RegisterAs(name2, validNoopFn)

				is.Equal(len(s.Errors()), 1)
				is.True(errors.Is(s.Errors()[0], hoist.ErrInvalidReturnNumber))

				expected := &hoist.ExportedService{
					Name: serviceName,
					Functions: map[string]*hoist.ExportedFunction{
						name2: {
							Name: name2,
						},
					},
				}
				is.Equal(s.Export(), expected)
			},
		},
		{
			Name: "invalid function: does not implement error type",
			Check: func(is *is.I, s *hoist.Service) {
				name := "myFn"
				name2 := "myFn2"

				badFn := func(*MyCtx, *MyParams) (*MyData, *MyData) {
					return nil, nil
				}
				s.RegisterAs(name, badFn)
				s.RegisterAs(name2, validNoopFn)

				is.Equal(len(s.Errors()), 1)
				is.True(errors.Is(s.Errors()[0], hoist.ErrInvalidReturnMissingError))

				expected := &hoist.ExportedService{
					Name: serviceName,
					Functions: map[string]*hoist.ExportedFunction{
						name2: {
							Name: name2,
						},
					},
				}
				is.Equal(s.Export(), expected)
			},
		},
		{
			Name: "multiple errors",
			Check: func(is *is.I, s *hoist.Service) {
				name := "myFn"
				name2 := "myFn2"

				badFn := func(*MyCtx, *MyParams) (*MyData, *MyData) {
					return nil, nil
				}
				badFn2 := func(*MyCtx) (*MyData, error) {
					return nil, nil
				}
				s.RegisterAs(name, badFn)
				s.RegisterAs(name2, badFn2)

				is.Equal(len(s.Errors()), 2)
				is.True(errors.Is(s.Errors()[0], hoist.ErrInvalidReturnMissingError))
				is.True(errors.Is(s.Errors()[0], hoist.ErrInvalidFunction))
				is.True(errors.Is(s.Errors()[1], hoist.ErrInvalidParameterNumber))
				is.True(errors.Is(s.Errors()[1], hoist.ErrInvalidFunction))

				expected := &hoist.ExportedService{
					Name:      serviceName,
					Functions: map[string]*hoist.ExportedFunction{},
				}
				is.Equal(s.Export(), expected)
			},
		},
	}

	for _, entry := range table {
		t.Run(entry.Name, func(t *testing.T) {
			is := is.New(t)

			s := hoist.NewService(serviceName)
			entry.Check(is, s)
		})
	}
}

type MyCtx struct{}
type MyParams struct{}
type MyData struct{}

func validNoopFn(*MyCtx, *MyParams) (*MyData, error) {
	return nil, nil
}

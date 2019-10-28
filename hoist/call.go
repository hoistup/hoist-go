package hoist

import (
	"github.com/JosiahWitt/erk"
)

type ErkFunctionNotFound erk.DefaultKind

var (
	ErrFunctionNotFound   = erk.New(ErkFunctionNotFound{}, "service '{{.serviceName}}' does not have function '{{.fnName}}'")
	ErrFunctionCallFailed = erk.New(ErkFunctionCall{}, "service '{{.serviceName}}': error while calling function '{{.fnName}}': {{.err}}")
)

// Call a function, providing the function name and JSON encoded rawParams.
func (s *Service) Call(fnName string, rawParams []byte) (interface{}, error) {
	s.mu.RLock()
	fn, ok := s.funcs[fnName]
	if !ok {
		return nil, erk.WithParams(ErrFunctionNotFound, erk.Params{"serviceName": s.name, "fnName": fnName})
	}
	s.mu.RUnlock()

	data, err := fn(rawParams)
	if err != nil {
		return nil, erk.WrapAs(erk.WithParams(ErrFunctionCallFailed, erk.Params{"serviceName": s.name, "fnName": fnName}), err)
	}

	return data, nil
}

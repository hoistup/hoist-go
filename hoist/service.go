// Package hoist provides the SDK to build a hoist server application in Go.
package hoist

import (
	"sync"
)

// rawFunc is an internal representation of a registered function.
type rawFunc func(rawParams []byte) (interface{}, error)

// Service represents a server instance of a hoist application.
type Service struct {
	mu sync.RWMutex

	name   string
	errors []error

	funcs map[string]rawFunc
}

// NewService creates a new service with the provided name.
func NewService(name string) *Service {
	return &Service{
		name:  name,
		funcs: make(map[string]rawFunc),
	}
}

// Errors returns all the errors associated with the service.
func (s *Service) Errors() []error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	eCopy := make([]error, len(s.errors))
	copy(eCopy, s.errors)

	return eCopy
}

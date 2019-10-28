package hoist

// ExportedFunction with name, parameters, and return values.
type ExportedFunction struct {
	Name string `json:"name"`
}

// ExportedService with name and functions.
type ExportedService struct {
	Name      string                       `json:"name"`
	Functions map[string]*ExportedFunction `json:"functions"`
}

// Export the service into a static representation of the API.
func (s *Service) Export() *ExportedService {
	s.mu.RLock()
	defer s.mu.RUnlock()

	service := ExportedService{
		Name:      s.name,
		Functions: make(map[string]*ExportedFunction),
	}

	// Build up the functions
	// TODO: Provide information about the parameters and return types
	for name := range s.funcs {
		service.Functions[name] = &ExportedFunction{
			Name: name,
		}
	}

	return &service
}

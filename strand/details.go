// Package strand defines strands that are encoded with wire.
package strand

// RequestDetails are the details encoded with wire for a request.
type RequestDetails struct {
	RequestID    string `json:"id"`
	ServiceName  string `json:"svc"`
	FunctionName string `json:"fn"`
}

// ResponseDetails are the details encoded with wire for a response.
type ResponseDetails struct {
	RequestID       string `json:"id"`
	IsError         bool   `json:"err,omitempty"`
	IsInternalError bool   `json:"ierr,omitempty"`
}

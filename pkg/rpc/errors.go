package rpc

import (
	"fmt"
)

type RequestError struct {
	Method     string
	StatusCode int
}

func (r *RequestError) Error() string {
	return fmt.Sprintf("%s - wrong status: %d", r.Method, r.StatusCode)
}

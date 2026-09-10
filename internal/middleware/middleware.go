package middleware

import "net/http"

// Middleware
type Middleware func(http.Handler) http.Handler

// Apply applies a middleware to an http handler func and returns an http handler.
func (m Middleware) Apply(f http.HandlerFunc) http.Handler {
	return m(http.HandlerFunc(f))
}

package middleware

import "net/http"

// CSRF is a middleware for preventing CSRF attacks.
func CSRF(handler http.Handler) http.Handler {
	cop := http.NewCrossOriginProtection()

	cop.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("CSRF check failed"))
	}))

	return cop.Handler(handler)

}

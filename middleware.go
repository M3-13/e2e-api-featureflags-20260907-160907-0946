package main

import (
	"log"
	"net/http"
)

// withLogging wraps next with access logging. The full logging behaviour is
// implemented by the dedicated middleware ticket; this scaffold only forwards
// the request unchanged.
func withLogging(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

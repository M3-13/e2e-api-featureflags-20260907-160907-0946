package main

import (
	"log"
	"net/http"
	"strings"
	"time"
	"unicode"
)

// statusRecorder wraps an http.ResponseWriter to capture the response status
// code. The status defaults to 200 (http.StatusOK) when the handler never
// writes one explicitly.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

// sanitize removes all control characters (such as \r and \n) from s so a
// single request cannot forge additional log entries.
func sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}

// withLogging wraps next with access logging. After the handler returns it
// writes exactly one log line containing the method, the path (without query
// string), the status code and the request duration. Only these four values
// are ever logged; the query string and request body are never read or logged.
func withLogging(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w}

		start := time.Now()
		next.ServeHTTP(rec, r)
		duration := time.Since(start)

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}

		logger.Printf("%s %s %d %s",
			sanitize(r.Method),
			sanitize(r.URL.Path),
			status,
			duration)
	})
}

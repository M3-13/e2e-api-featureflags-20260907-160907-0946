package main

import (
	"encoding/json"
	"net/http"
)

// maxBodyBytes is the fixed request body size limit (1 MiB).
const maxBodyBytes = 1 << 20

// writeJSON writes v as a JSON response with the given HTTP status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error object of the shape {"error": msg}.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// decodeJSON decodes the request body into dst, enforcing a 1 MiB size limit.
// Unknown JSON fields are ignored. It returns an error if the body is larger
// than the limit or is not valid JSON.
func decodeJSON(r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodyBytes)
	return json.NewDecoder(r.Body).Decode(dst)
}

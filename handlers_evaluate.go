package main

import "net/http"

// evaluateFlagHandler handles GET /flags/{key}/evaluate.
func (s *server) evaluateFlagHandler(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

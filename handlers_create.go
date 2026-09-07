package main

import "net/http"

// createFlagHandler handles POST /flags.
func (s *server) createFlagHandler(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

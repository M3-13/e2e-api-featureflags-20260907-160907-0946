package main

import "net/http"

// updateFlagHandler handles PUT /flags/{key}.
func (s *server) updateFlagHandler(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

// deleteFlagHandler handles DELETE /flags/{key}.
func (s *server) deleteFlagHandler(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

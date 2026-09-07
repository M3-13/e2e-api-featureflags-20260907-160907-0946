package main

import "net/http"

// listFlagsHandler handles GET /flags.
func (s *server) listFlagsHandler(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

// getFlagHandler handles GET /flags/{key}.
func (s *server) getFlagHandler(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

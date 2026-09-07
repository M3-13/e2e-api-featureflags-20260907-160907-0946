package main

import "net/http"

// listFlagsHandler handles GET /flags. It answers 200 with the JSON array of
// all stored feature flags (an empty array when the store is empty).
func (s *server) listFlagsHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.List())
}

// getFlagHandler handles GET /flags/{key}. It answers 200 with the feature on
// a known key, and 404 with a JSON error object on an unknown key.
func (s *server) getFlagHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	flag, ok := s.store.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}
	writeJSON(w, http.StatusOK, flag)
}

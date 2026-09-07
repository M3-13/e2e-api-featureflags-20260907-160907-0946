package main

import "net/http"

// updateFlagRequest is the JSON body accepted by PUT /flags/{key}. It mirrors
// the POST /flags body minus the key, which is immutable and taken from the
// path. Enabled is required (a missing enabled is indistinguishable from false
// without a pointer), Description is optional and RolloutPercent must be 0-100.
type updateFlagRequest struct {
	Enabled        *bool  `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent int    `json:"rollout_percent"`
}

// updateFlagHandler handles PUT /flags/{key}.
func (s *server) updateFlagHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var req updateFlagRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Enabled == nil {
		writeError(w, http.StatusBadRequest, "enabled is required")
		return
	}
	if req.RolloutPercent < 0 || req.RolloutPercent > 100 {
		writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
		return
	}

	updated := Feature{
		Key:            key,
		Enabled:        *req.Enabled,
		Description:    req.Description,
		RolloutPercent: req.RolloutPercent,
	}

	f, ok := s.store.Update(key, updated)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}

	writeJSON(w, http.StatusOK, f)
}

// deleteFlagHandler handles DELETE /flags/{key}.
func (s *server) deleteFlagHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if !s.store.Delete(key) {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

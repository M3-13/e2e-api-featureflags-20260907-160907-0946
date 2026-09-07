package main

import (
	"net/http"
	"strings"
)

// createFlagRequest is the JSON body accepted by POST /flags. Enabled is a
// *bool so that a missing field can be distinguished from an explicit false.
type createFlagRequest struct {
	Key            string `json:"key"`
	Enabled        *bool  `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent int    `json:"rollout_percent"`
}

// createFlagHandler handles POST /flags. It validates the request body,
// creates the flag and answers 201 with the stored flag. A duplicate key
// answers 409; an invalid or oversized body answers 400.
func (s *server) createFlagHandler(w http.ResponseWriter, r *http.Request) {
	var req createFlagRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Key) == "" {
		writeError(w, http.StatusBadRequest, "key is required")
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

	feature := Feature{
		Key:            req.Key,
		Enabled:        *req.Enabled,
		Description:    req.Description,
		RolloutPercent: req.RolloutPercent,
	}

	if err := s.store.Create(feature); err != nil {
		if err == ErrExists {
			writeError(w, http.StatusConflict, "feature already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create feature")
		return
	}

	writeJSON(w, http.StatusCreated, feature)
}

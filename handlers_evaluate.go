package main

import (
	"hash/fnv"
	"net/http"
)

// evaluateFlagHandler handles GET /flags/{key}/evaluate?user=ID.
//
// The decision is deterministic for a given key and user: an FNV-1a hash of
// the key and user (joined with a separator) is reduced modulo 100 and
// compared against the flag's rollout_percent. The user is used only for the
// hash and is never stored.
func (s *server) evaluateFlagHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	user := r.URL.Query().Get("user")

	if user == "" {
		writeError(w, http.StatusBadRequest, "missing user parameter")
		return
	}

	flag, ok := s.store.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "feature not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"decision": evaluate(flag, key, user)})
}

// evaluate computes the rollout decision for a flag, key and user without
// touching the store. It is pure and deterministic.
func evaluate(f Feature, key, user string) bool {
	if !f.Enabled {
		return false
	}
	if f.RolloutPercent <= 0 {
		return false
	}
	if f.RolloutPercent >= 100 {
		return true
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	_, _ = h.Write([]byte(":"))
	_, _ = h.Write([]byte(user))

	return int(h.Sum32()%100) < f.RolloutPercent
}

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func mutateTestMux(store *Store) http.Handler {
	return newMux(store, log.New(os.Stderr, "", 0))
}

func seedFlag(t *testing.T, store *Store, f Feature) {
	t.Helper()
	if err := store.Create(f); err != nil {
		t.Fatalf("Create(%q): unexpected error %v", f.Key, err)
	}
}

func TestUpdateFlag(t *testing.T) {
	store := NewStore()
	seedFlag(t, store, Feature{Key: "f1", Enabled: true, Description: "old", RolloutPercent: 50})

	body := `{"enabled": false, "description": "new", "rollout_percent": 25}`
	req := httptest.NewRequest(http.MethodPut, "/flags/f1", strings.NewReader(body))
	rec := httptest.NewRecorder()
	mutateTestMux(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	var got Feature
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unexpected non-JSON body %q: %v", rec.Body.String(), err)
	}
	if got.Key != "f1" {
		t.Fatalf("expected key %q, got %q", "f1", got.Key)
	}
	if got.Enabled {
		t.Fatalf("expected enabled false, got true")
	}
	if got.Description != "new" {
		t.Fatalf("expected description %q, got %q", "new", got.Description)
	}
	if got.RolloutPercent != 25 {
		t.Fatalf("expected rollout_percent 25, got %d", got.RolloutPercent)
	}

	stored, ok := store.Get("f1")
	if !ok {
		t.Fatalf("expected flag to still exist after update")
	}
	if stored.Description != "new" || stored.RolloutPercent != 25 || stored.Enabled {
		t.Fatalf("store not updated: %+v", stored)
	}
}

func TestUpdateFlagPreservesKey(t *testing.T) {
	store := NewStore()
	seedFlag(t, store, Feature{Key: "f1", Enabled: true})

	body := `{"enabled": false, "key": "hijacked"}`
	req := httptest.NewRequest(http.MethodPut, "/flags/f1", strings.NewReader(body))
	rec := httptest.NewRecorder()
	mutateTestMux(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	var got Feature
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unexpected non-JSON body %q: %v", rec.Body.String(), err)
	}
	if got.Key != "f1" {
		t.Fatalf("key must remain unchanged, got %q", got.Key)
	}
	if _, ok := store.Get("hijacked"); ok {
		t.Fatalf("key must not be changed to %q", "hijacked")
	}
}

func TestUpdateFlagUnknownKey(t *testing.T) {
	store := NewStore()

	body := `{"enabled": true}`
	req := httptest.NewRequest(http.MethodPut, "/flags/missing", strings.NewReader(body))
	rec := httptest.NewRecorder()
	mutateTestMux(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestUpdateFlagInvalidBody(t *testing.T) {
	store := NewStore()
	seedFlag(t, store, Feature{Key: "f1", Enabled: true})

	tests := []struct {
		name string
		body string
	}{
		{"missing enabled", `{"description": "x"}`},
		{"malformed json", `{not json`},
		{"rollout below range", `{"enabled": true, "rollout_percent": -1}`},
		{"rollout above range", `{"enabled": true, "rollout_percent": 101}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/flags/f1", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			mutateTestMux(store).ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d (%s)", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestDeleteFlag(t *testing.T) {
	store := NewStore()
	seedFlag(t, store, Feature{Key: "f1", Enabled: true})

	req := httptest.NewRequest(http.MethodDelete, "/flags/f1", nil)
	rec := httptest.NewRecorder()
	mutateTestMux(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", rec.Body.String())
	}

	if _, ok := store.Get("f1"); ok {
		t.Fatalf("expected flag to be gone after delete")
	}
}

func TestDeleteFlagUnknownKey(t *testing.T) {
	store := NewStore()

	req := httptest.NewRequest(http.MethodDelete, "/flags/missing", nil)
	rec := httptest.NewRecorder()
	mutateTestMux(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

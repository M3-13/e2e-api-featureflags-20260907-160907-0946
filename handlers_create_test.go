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

func newCreateTestServer() (*server, *Store) {
	store := NewStore()
	logger := log.New(os.Stderr, "", 0)
	return &server{store: store, logger: logger}, store
}

func postFlags(s *server, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rec := httptest.NewRecorder()
	s.createFlagHandler(rec, req)
	return rec
}

func TestCreateFlagValidBody(t *testing.T) {
	s, store := newCreateTestServer()
	rec := postFlags(s, `{"key":"beta","enabled":true,"description":"beta flag","rollout_percent":50}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d (body %q)", rec.Code, rec.Body.String())
	}

	var got Feature
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unexpected non-JSON body %q: %v", rec.Body.String(), err)
	}

	want := Feature{Key: "beta", Enabled: true, Description: "beta flag", RolloutPercent: 50}
	if got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}

	stored, ok := store.Get("beta")
	if !ok {
		t.Fatalf("expected flag to be stored")
	}
	if stored != want {
		t.Fatalf("expected stored flag %+v, got %+v", want, stored)
	}
}

func TestCreateFlagDefaultsRolloutPercent(t *testing.T) {
	s, _ := newCreateTestServer()
	rec := postFlags(s, `{"key":"gamma","enabled":false}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d (body %q)", rec.Code, rec.Body.String())
	}

	var got Feature
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unexpected non-JSON body %q: %v", rec.Body.String(), err)
	}
	if got.RolloutPercent != 0 {
		t.Fatalf("expected default rollout_percent 0, got %d", got.RolloutPercent)
	}
	if got.Enabled {
		t.Fatalf("expected enabled false, got true")
	}
}

func TestCreateFlagBrokenJSON(t *testing.T) {
	s, _ := newCreateTestServer()
	rec := postFlags(s, `{"key": "broken"`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	assertErrorBody(t, rec)
}

func TestCreateFlagMissingKey(t *testing.T) {
	s, _ := newCreateTestServer()
	rec := postFlags(s, `{"enabled":true}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	assertErrorBody(t, rec)
}

func TestCreateFlagEmptyKey(t *testing.T) {
	s, _ := newCreateTestServer()
	rec := postFlags(s, `{"key":"","enabled":true}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	assertErrorBody(t, rec)
}

func TestCreateFlagMissingEnabled(t *testing.T) {
	s, _ := newCreateTestServer()
	rec := postFlags(s, `{"key":"alpha"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	assertErrorBody(t, rec)
}

func TestCreateFlagEnabledFalseIsValid(t *testing.T) {
	s, store := newCreateTestServer()
	rec := postFlags(s, `{"key":"alpha","enabled":false}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d (body %q)", rec.Code, rec.Body.String())
	}
	stored, ok := store.Get("alpha")
	if !ok {
		t.Fatalf("expected flag to be stored")
	}
	if stored.Enabled {
		t.Fatalf("expected enabled false, got true")
	}
}

func TestCreateFlagRolloutPercentOutOfRange(t *testing.T) {
	for _, pct := range []string{"-1", "101"} {
		s, _ := newCreateTestServer()
		rec := postFlags(s, `{"key":"alpha","enabled":true,"rollout_percent":`+pct+`}`)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("rollout_percent %s: expected status 400, got %d", pct, rec.Code)
		}
		assertErrorBody(t, rec)
	}
}

func TestCreateFlagDuplicateKey(t *testing.T) {
	s, _ := newCreateTestServer()
	body := `{"key":"dup","enabled":true}`

	if rec := postFlags(s, body); rec.Code != http.StatusCreated {
		t.Fatalf("first create: expected 201, got %d", rec.Code)
	}

	rec := postFlags(s, body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate create: expected status 409, got %d", rec.Code)
	}
	assertErrorBody(t, rec)
}

func TestCreateFlagBodyTooLarge(t *testing.T) {
	s, _ := newCreateTestServer()
	big := strings.Repeat("a", maxBodyBytes+1)
	body := `{"key":"big","enabled":true,"description":"` + big + `"}`

	rec := postFlags(s, body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	assertErrorBody(t, rec)
}

func TestCreateFlagDiscardsUnknownFields(t *testing.T) {
	s, store := newCreateTestServer()
	body := `{"key":"delta","enabled":true,"unknown":"ignored","secret":"value"}`

	rec := postFlags(s, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d (body %q)", rec.Code, rec.Body.String())
	}

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unexpected non-JSON body %q: %v", rec.Body.String(), err)
	}
	if _, ok := got["unknown"]; ok {
		t.Fatalf("unknown field leaked into response: %+v", got)
	}
	if _, ok := got["secret"]; ok {
		t.Fatalf("unknown field leaked into response: %+v", got)
	}

	stored, ok := store.Get("delta")
	if !ok {
		t.Fatalf("expected flag to be stored")
	}
	if stored.Description != "" {
		t.Fatalf("expected empty description, got %q", stored.Description)
	}
}

func assertErrorBody(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON error object, got %q: %v", rec.Body.String(), err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatalf("expected error object with \"error\" field, got %+v", body)
	}
}

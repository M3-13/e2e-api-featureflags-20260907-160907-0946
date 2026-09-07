package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newReadTestServer() *server {
	logger := log.New(os.Stderr, "", 0)
	return &server{store: NewStore(), logger: logger}
}

func TestListFlagsEmpty(t *testing.T) {
	s := newReadTestServer()
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rec := httptest.NewRecorder()

	s.listFlagsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got []Feature
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("got %v, want empty array", got)
	}
}

func TestListFlagsMultiple(t *testing.T) {
	s := newReadTestServer()
	flags := []Feature{
		{Key: "alpha", Enabled: true, Description: "first", RolloutPercent: 50},
		{Key: "beta", Enabled: false, Description: "second", RolloutPercent: 0},
		{Key: "gamma", Enabled: true, Description: "", RolloutPercent: 100},
	}
	for _, f := range flags {
		if err := s.store.Create(f); err != nil {
			t.Fatalf("Create(%s): %v", f.Key, err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rec := httptest.NewRecorder()
	s.listFlagsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got []Feature
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(got) != len(flags) {
		t.Fatalf("len = %d, want %d", len(got), len(flags))
	}
}

func TestGetFlagExistingKey(t *testing.T) {
	s := newReadTestServer()
	flag := Feature{Key: "alpha", Enabled: true, Description: "desc", RolloutPercent: 25}
	if err := s.store.Create(flag); err != nil {
		t.Fatalf("Create: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/flags/alpha", nil)
	req.SetPathValue("key", "alpha")
	rec := httptest.NewRecorder()
	s.getFlagHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got Feature
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got != flag {
		t.Fatalf("got %+v, want %+v", got, flag)
	}
}

func TestGetFlagUnknownKey(t *testing.T) {
	s := newReadTestServer()

	req := httptest.NewRequest(http.MethodGet, "/flags/missing", nil)
	req.SetPathValue("key", "missing")
	rec := httptest.NewRecorder()
	s.getFlagHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := got["error"]; !ok {
		t.Fatalf("response %v missing \"error\" field", got)
	}
}

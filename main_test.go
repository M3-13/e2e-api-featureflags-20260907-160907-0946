package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func testMux() http.Handler {
	return newMux(NewStore(), log.New(os.Stderr, "", 0))
}

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	testMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected non-JSON body %q: %v", rec.Body.String(), err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", body["status"])
	}
}

func TestNoCORSHeaders(t *testing.T) {
	for _, target := range []string{"/healthz", "/flags", "/flags/x", "/flags/x/evaluate"} {
		method := http.MethodGet
		if target == "/flags" {
			method = http.MethodPost
		}
		req := httptest.NewRequest(method, target, nil)
		rec := httptest.NewRecorder()
		testMux().ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("%s %s: unexpected Access-Control-Allow-Origin header %q", method, target, got)
		}
		if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "" {
			t.Fatalf("%s %s: unexpected Access-Control-Allow-Methods header %q", method, target, got)
		}
		if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "" {
			t.Fatalf("%s %s: unexpected Access-Control-Allow-Headers header %q", method, target, got)
		}
	}
}

func TestStoreConcurrency(t *testing.T) {
	store := NewStore()

	const goroutines = 100
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "flag-" + string(rune('a'+i%26)) + string(rune('a'+i))
			f := Feature{
				Key:            key,
				Enabled:        i%2 == 0,
				Description:    "concurrent feature",
				RolloutPercent: i % 101,
			}
			if err := store.Create(f); err != nil && err != ErrExists {
				t.Errorf("Create(%q): unexpected error %v", key, err)
				return
			}
			_ = store.List()
			if _, ok := store.Get(key); !ok {
				t.Errorf("Get(%q): expected flag to exist", key)
			}
			if _, ok := store.Update(key, f); !ok {
				t.Errorf("Update(%q): expected flag to exist", key)
			}
			_ = store.List()
			if !store.Delete(key) {
				t.Errorf("Delete(%q): expected flag to exist", key)
			}
		}(i)
	}
	wg.Wait()
}

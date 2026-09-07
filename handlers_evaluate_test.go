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

func evaluateTestMux(store *Store) http.Handler {
	return newMux(store, log.New(os.Stderr, "", 0))
}

func evaluateDecision(t *testing.T, store *Store, key, user string) (int, bool) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/flags/"+key+"/evaluate?user="+user, nil)
	rec := httptest.NewRecorder()
	evaluateTestMux(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		return rec.Code, false
	}
	var body map[string]bool
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected non-JSON body %q: %v", rec.Body.String(), err)
	}
	decision, ok := body["decision"]
	if !ok {
		t.Fatalf("response missing decision field: %q", rec.Body.String())
	}
	return rec.Code, decision
}

func TestEvaluateDeterministic(t *testing.T) {
	store := NewStore()
	store.Create(Feature{Key: "f", Enabled: true, RolloutPercent: 50})

	for i := 0; i < 20; i++ {
		_, first := evaluateDecision(t, store, "f", "user-1")
		for j := 0; j < 5; j++ {
			_, again := evaluateDecision(t, store, "f", "user-1")
			if again != first {
				t.Fatalf("decision for key=f user=user-1 changed across calls: %v then %v", first, again)
			}
		}
	}
}

func TestEvaluateRolloutZeroAlwaysFalse(t *testing.T) {
	store := NewStore()
	store.Create(Feature{Key: "f", Enabled: true, RolloutPercent: 0})

	for _, user := range []string{"a", "b", "c", "d"} {
		_, decision := evaluateDecision(t, store, "f", user)
		if decision {
			t.Fatalf("rollout_percent 0 should always be false, got true for user %q", user)
		}
	}
}

func TestEvaluateRolloutHundredAlwaysTrue(t *testing.T) {
	store := NewStore()
	store.Create(Feature{Key: "f", Enabled: true, RolloutPercent: 100})

	for _, user := range []string{"a", "b", "c", "d"} {
		_, decision := evaluateDecision(t, store, "f", user)
		if !decision {
			t.Fatalf("rollout_percent 100 should always be true (enabled), got false for user %q", user)
		}
	}
}

func TestEvaluateDisabledAlwaysFalse(t *testing.T) {
	store := NewStore()
	store.Create(Feature{Key: "f", Enabled: false, RolloutPercent: 100})

	_, decision := evaluateDecision(t, store, "f", "user")
	if decision {
		t.Fatalf("enabled=false should always be false regardless of rollout_percent")
	}
}

func TestEvaluateMissingUserBadRequest(t *testing.T) {
	store := NewStore()
	store.Create(Feature{Key: "f", Enabled: true, RolloutPercent: 100})

	req := httptest.NewRequest(http.MethodGet, "/flags/f/evaluate", nil)
	rec := httptest.NewRecorder()
	evaluateTestMux(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing user, got %d", rec.Code)
	}
}

func TestEvaluateUnknownKeyNotFound(t *testing.T) {
	store := NewStore()

	req := httptest.NewRequest(http.MethodGet, "/flags/nope/evaluate?user=u", nil)
	rec := httptest.NewRecorder()
	evaluateTestMux(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown key, got %d", rec.Code)
	}
}

func TestEvaluateUserNotStored(t *testing.T) {
	store := NewStore()
	store.Create(Feature{Key: "f", Enabled: true, RolloutPercent: 100})

	_, _ = evaluateDecision(t, store, "f", "secret-user-123")

	flags := store.List()
	if len(flags) != 1 {
		t.Fatalf("store should still contain exactly the created flag, got %d entries", len(flags))
	}
	if strings.Contains(flags[0].Key, "secret-user-123") {
		t.Fatalf("user leaked into stored flag key: %q", flags[0].Key)
	}
}

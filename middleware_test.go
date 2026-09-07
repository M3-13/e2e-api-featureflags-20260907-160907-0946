package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// logHandler returns a handler that answers the given status code.
func logHandler(status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	})
}

func runLogged(t *testing.T, logger *log.Logger, method, target string, next http.Handler) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	withLogging(next, logger).ServeHTTP(rec, req)
	return rec
}

func TestWithLoggingSingleLineFourFields(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	runLogged(t, logger, http.MethodGet, "/healthz", logHandler(http.StatusOK))

	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected exactly one log line, got %d: %q", len(lines), out)
	}

	fields := strings.Fields(lines[0])
	if len(fields) != 4 {
		t.Fatalf("expected four fields (method, path, status, duration), got %d: %q", len(fields), lines[0])
	}
	if fields[0] != http.MethodGet {
		t.Fatalf("expected method %q, got %q", http.MethodGet, fields[0])
	}
	if fields[1] != "/healthz" {
		t.Fatalf("expected path %q, got %q", "/healthz", fields[1])
	}
	if fields[2] != "200" {
		t.Fatalf("expected status %q, got %q", "200", fields[2])
	}
	if fields[3] == "" {
		t.Fatalf("expected a non-empty duration field: %q", lines[0])
	}
}

func TestWithLoggingStripsQueryString(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	runLogged(t, logger, http.MethodGet, "/flags/abc/evaluate?user=alice", logHandler(http.StatusOK))

	out := buf.String()
	if strings.Contains(out, "?") {
		t.Fatalf("query string must not be logged: %q", out)
	}
	if !strings.Contains(out, "/flags/abc/evaluate") {
		t.Fatalf("expected path without query, got %q", out)
	}
	if strings.Contains(out, "alice") {
		t.Fatalf("user query parameter must not be logged: %q", out)
	}
}

func TestWithLoggingSanitizesControlCharacters(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	// httptest.NewRequest rejects control characters, so build the request
	// directly to exercise the sanitisation path.
	req := &http.Request{
		Method: "GE\r\nT",
		URL:    &url.URL{Path: "/healt\r\nhz"},
		Header: make(http.Header),
	}
	rec := httptest.NewRecorder()
	withLogging(logHandler(http.StatusOK), logger).ServeHTTP(rec, req)

	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("control characters must not produce additional lines, got %d: %q", len(lines), out)
	}
	for _, c := range []string{"\r", "\n"} {
		if strings.Contains(lines[0], c) {
			t.Fatalf("log line still contains control character %q: %q", c, lines[0])
		}
	}
	if !strings.Contains(lines[0], "GET") {
		t.Fatalf("expected sanitized method GET, got %q", lines[0])
	}
}

func TestWithLoggingCapturesWrappedStatus(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	runLogged(t, logger, http.MethodGet, "/flags/missing", logHandler(http.StatusNotFound))

	if !strings.Contains(buf.String(), "404") {
		t.Fatalf("expected logged status 404, got %q", buf.String())
	}
}

func TestWithLoggingDefaultStatus200(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	// A handler that writes nothing must still log status 200.
	runLogged(t, logger, http.MethodGet, "/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	if !strings.Contains(buf.String(), "200") {
		t.Fatalf("expected default status 200, got %q", buf.String())
	}
}

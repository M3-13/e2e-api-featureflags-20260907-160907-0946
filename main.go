package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

// server holds the dependencies shared by all handlers.
type server struct {
	store  *Store
	logger *log.Logger
}

// newMux builds the full routing table and wraps it with the logging
// middleware.
func newMux(store *Store, logger *log.Logger) http.Handler {
	s := &server{store: store, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /flags", s.createFlagHandler)
	mux.HandleFunc("GET /flags", s.listFlagsHandler)
	mux.HandleFunc("GET /flags/{key}", s.getFlagHandler)
	mux.HandleFunc("PUT /flags/{key}", s.updateFlagHandler)
	mux.HandleFunc("DELETE /flags/{key}", s.deleteFlagHandler)
	mux.HandleFunc("GET /flags/{key}/evaluate", s.evaluateFlagHandler)
	mux.HandleFunc("GET /healthz", s.healthzHandler)
	return withLogging(mux, logger)
}

// healthzHandler answers 200 {"status":"ok"}.
func (s *server) healthzHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	store := NewStore()

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           newMux(store, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Printf("feature flag service listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("server error: %v", err)
	}
}

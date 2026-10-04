// Package web is the HTTP delivery adapter. It serves the landing page at
// the same domain as the SSH portfolio, teaching visitors how to connect:
//
//	browser  (Accept: text/html)  ->  HTML landing page with a live preview
//	curl / wget / anything else   ->  a colored terminal card
//
// HTTPS is handled in front of this server by Caddy (see deploy/), so this
// package only speaks plain HTTP on an internal port.
package web

import (
	"context"
	"errors"
	"net/http"
	"time"

	"charm.land/log/v2"

	"ssh-portfolio/internal/domain"
)

// Portfolio is the read-only content the web pages need. Defined here, where
// it's consumed; usecase.PortfolioService implements it.
type Portfolio interface {
	Site() domain.Site
	Profile() domain.Profile
	Experiences() []domain.Experience
	Projects() []domain.Project
	Skills() []domain.SkillGroup
	Contacts() []domain.Contact
	YearsOfExperience() int
}

// Config holds the web server's settings.
type Config struct {
	Addr string // listen address, e.g. ":8080"

	// SSHCommand is what visitors should type, e.g. "ssh rivaldo.dev".
	SSHCommand string

	// Fingerprint is the SSH host key fingerprint ("SHA256:..."), shown so
	// visitors can verify the server on first connect. Optional.
	Fingerprint string
}

// Server serves the landing page.
type Server struct {
	cfg  Config
	data Portfolio
	srv  *http.Server
}

// New builds the web server.
func New(cfg Config, data Portfolio) *Server {
	s := &Server{cfg: cfg, data: data}
	s.srv = &http.Server{
		Addr:              cfg.Addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

// Handler returns the HTTP routes. Exposed for tests.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok\n"))
	})
	return securityHeaders(mux)
}

// Run serves until ctx is cancelled, then shuts down gracefully.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		log.Info("starting web server", "addr", s.cfg.Addr)
		errCh <- s.srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

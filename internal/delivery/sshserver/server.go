// Package sshserver is the SSH delivery adapter: it accepts SSH connections
// with Wish and runs a Bubble Tea program for each session.
//
// It doesn't know what the program shows. The composition root passes in a
// ModelFactory, so the same server could serve any TUI.
package sshserver

import (
	"context"
	"errors"
	"net"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"charm.land/ssh"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"
	"charm.land/wish/v2/logging"
	gossh "golang.org/x/crypto/ssh"
)

// Config holds the server's settings.
type Config struct {
	Host        string
	Port        string
	HostKeyPath string // generated on first run if missing; keep it stable in production
	IdleTimeout time.Duration
}

// ModelFactory creates a fresh TUI model for each SSH session.
type ModelFactory func(sess ssh.Session) tea.Model

// Server wraps a configured Wish server.
type Server struct {
	srv *ssh.Server
}

// New builds a server that runs newModel for every session.
func New(cfg Config, newModel ModelFactory) (*Server, error) {
	srv, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(cfg.Host, cfg.Port)),
		wish.WithHostKeyPath(cfg.HostKeyPath),
		// Public portfolio: let everyone in, with or without an SSH key.
		wish.WithPublicKeyAuth(func(ssh.Context, ssh.PublicKey) bool { return true }),
		wish.WithKeyboardInteractiveAuth(
			func(ssh.Context, gossh.KeyboardInteractiveChallenge) bool { return true },
		),
		wish.WithIdleTimeout(cfg.IdleTimeout),
		// Middleware runs bottom-up: logging -> activeterm -> bubbletea.
		wish.WithMiddleware(
			bubbletea.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
				return newModel(s), nil
			}),
			activeterm.Middleware(), // require a PTY (rejects `ssh host some-command`)
			logging.Middleware(),
		),
	)
	if err != nil {
		return nil, err
	}
	return &Server{srv: srv}, nil
}

// Fingerprint returns the host key's SHA256 fingerprint, the value SSH
// clients show on first connect ("SHA256:..."). Publishing it lets visitors
// verify they're talking to the real server.
func (s *Server) Fingerprint() string {
	if len(s.srv.HostSigners) == 0 {
		return ""
	}
	return gossh.FingerprintSHA256(s.srv.HostSigners[0].PublicKey())
}

// Run serves until ctx is cancelled, then shuts down gracefully, giving
// open sessions up to 30 seconds to finish.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		log.Info("starting SSH server", "addr", s.srv.Addr)
		errCh <- s.srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		return err
	}
	return nil
}

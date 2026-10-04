// Command ssh-portfolio serves Rivaldo Ardika Lawalata's portfolio over SSH.
//
// This is the composition root: the only place that knows about every
// layer. It reads flags, builds the concrete adapters, and plugs them
// together. See docs/ARCHITECTURE.md.
//
//	go run ./cmd/ssh-portfolio            # SSH on :23234
//	go run ./cmd/ssh-portfolio -local     # TUI in this terminal, no servers
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/ssh"

	"ssh-portfolio/content"
	"ssh-portfolio/internal/delivery/sshserver"
	"ssh-portfolio/internal/delivery/tui"
	"ssh-portfolio/internal/repository/yamlrepo"
	"ssh-portfolio/internal/usecase"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	host := flag.String("host", "0.0.0.0", "address to listen on")
	port := flag.String("port", "23234", "port to listen on (22 in production)")
	keyPath := flag.String("key", ".ssh/id_ed25519", "host key path (generated if missing)")
	contentPath := flag.String(
		"content",
		"",
		"portfolio YAML file (default: the embedded content/portfolio.yaml)",
	)
	local := flag.Bool("local", false, "run the TUI in this terminal instead of serving SSH")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Repository adapter: where content comes from.
	repo := yamlrepo.New(content.Default)
	if *contentPath != "" {
		var err error
		if repo, err = yamlrepo.FromFile(*contentPath); err != nil {
			return err
		}
	}

	// Use case: load, validate, and prepare content once at startup.
	svc, err := usecase.NewPortfolioService(ctx, repo, time.Now)
	if err != nil {
		return err
	}

	// Delivery adapters: how content reaches people.
	if *local {
		_, err := tea.NewProgram(tui.New(svc)).Run()
		return err
	}

	sshSrv, err := sshserver.New(sshserver.Config{
		Host:        *host,
		Port:        *port,
		HostKeyPath: *keyPath,
		IdleTimeout: 10 * time.Minute,
	}, func(ssh.Session) tea.Model { return tui.New(svc) })
	if err != nil {
		return err
	}

	return sshSrv.Run(ctx)
}

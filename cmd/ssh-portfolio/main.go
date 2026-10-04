// Command ssh-portfolio serves Rivaldo Ardika Lawalata's portfolio over SSH,
// plus a web landing page on the same domain that explains how to connect.
//
// This is the composition root: the only place that knows about every
// layer. It reads flags, builds the concrete adapters, and plugs them
// together. See docs/ARCHITECTURE.md.
//
//	go run ./cmd/ssh-portfolio            # SSH on :23234, web on :8080
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

	"golang.org/x/sync/errgroup"

	tea "charm.land/bubbletea/v2"
	"charm.land/ssh"

	"ssh-portfolio/content"
	"ssh-portfolio/internal/delivery/sshserver"
	"ssh-portfolio/internal/delivery/tui"
	"ssh-portfolio/internal/delivery/web"
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
	httpAddr := flag.String("http", ":8080", "web landing page address (empty to disable)")
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

	// Run SSH and web side by side; if either fails, both shut down.
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return sshSrv.Run(ctx) })
	if *httpAddr != "" {
		webSrv := web.New(web.Config{
			Addr:        *httpAddr,
			SSHCommand:  sshCommand(svc.Site().Address, *port),
			Fingerprint: sshSrv.Fingerprint(),
		}, svc)
		g.Go(func() error { return webSrv.Run(ctx) })
	}
	return g.Wait()
}

// sshCommand is what the landing page tells visitors to type: the bare
// domain in production (port 22), or the local dev address otherwise.
func sshCommand(address, port string) string {
	if port == "22" {
		return "ssh " + address
	}
	return "ssh -p " + port + " localhost"
}

# ssh-portfolio

Rivaldo Ardika Lawalata's portfolio, served over SSH and laid out like a website.

```sh
ssh rivaldo.dev          # the portfolio, in your terminal
curl rivaldo.dev         # a terminal card with the command
open https://rivaldo.dev # landing page that teaches visitors how to connect
```

All three are served from **one domain**: port 22 is SSH, ports 80/443 are the web page (via Caddy).

```
╭──────────────────────────────────────────────────────────────╮
│   ● ● ●                 ssh rivaldo.dev/experience           │
│    Home   Experience   Projects   Skills   Contact           │
│   ────────────────────────────────────────────────────────   │
│   Experience                                                 │
│   › Fullstack Developer · PT Woori Finance Indonesia         │
│     Sep 2024 – Present · Full-time                           │
│     Backend Developer · Waterhub                             │
│     ...                                                      │
│   ↑/↓ select  •  enter open  •  ←/→ pages  •  q quit         │
╰──────────────────────────────────────────────────────────────╯
```

Built with Go and [Charm](https://charm.land)'s stack: Wish (SSH server), Bubble Tea (TUI),
Lip Gloss (styling), and Bubbles (viewport). The code follows clean architecture.

## Quick start

Requires Go 1.26+.

```sh
make run        # SSH on :23234 (ssh -p 23234 localhost) + web on http://localhost:8080
make local      # same UI straight in this terminal, no SSH (fastest while editing)
make test       # vet + all tests
make build      # static binary at bin/ssh-portfolio
make docker     # container image
```

Flags for `go run ./cmd/ssh-portfolio`:

| Flag | Default | Purpose |
|---|---|---|
| `-port` | `23234` | listen port (`22` in production) |
| `-host` | `0.0.0.0` | listen address |
| `-key` | `.ssh/id_ed25519` | server host key; generated if missing |
| `-http` | `:8080` | landing page address (empty disables it) |
| `-content` | embedded | load a different portfolio YAML file |
| `-local` | `false` | run the TUI locally instead of serving SSH |

## Updating the content

All the text lives in **[`content/portfolio.yaml`](content/portfolio.yaml)**, so updating it doesn't touch Go code.
The schema is in [docs/CONTENT.md](docs/CONTENT.md). The file is embedded into the binary at build time,
and `make test` checks that it's valid.

## Project layout

```
cmd/ssh-portfolio/        composition root: flags + wiring (main.go)
content/                  portfolio.yaml, embedded into the binary
internal/
  domain/                 entities and business rules (no dependencies)
  usecase/                application logic + Repository port
  repository/yamlrepo/    adapter: loads the domain from YAML
  delivery/tui/           adapter: Bubble Tea UI
  delivery/sshserver/     adapter: Wish SSH server
  delivery/web/           adapter: HTTP landing page (+ curl card)
deploy/Caddyfile          HTTPS reverse proxy for the landing page
docker-compose.yml        production stack: portfolio (:22) + caddy (:80/:443)
docs/                     architecture, content, development, deployment
```

## Documentation

- [Architecture](docs/ARCHITECTURE.md): the layers, the dependency rule, and how a request flows.
- [Content](docs/CONTENT.md): the YAML schema and editing tips.
- [Development](docs/DEVELOPMENT.md): adding pages or data sources, testing, and gotchas.
- [Deployment](docs/DEPLOYMENT.md): one domain, DNS, Docker Compose, and HTTPS on a VPS.

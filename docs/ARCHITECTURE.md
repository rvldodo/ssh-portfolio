# Architecture

The project follows **clean architecture**. The code is split into layers, and
dependencies only point *inward*: toward the business core, never away from it.

```
            ┌──────────────────────────────────────────────────────────────┐
            │  cmd/ssh-portfolio  (composition root: wires it all)         │
            └──────────────────────────────────────────────────────────────┘
               │               │               │                  │
               ▼               ▼               ▼                  ▼
   ┌─────────────────┐ ┌──────────────┐ ┌──────────────┐ ┌────────────────────┐
   │delivery/        │ │ delivery/tui │ │ delivery/web │ │ repository/yamlrepo│  adapters
   │sshserver (Wish) │ │ (Bubble Tea) │ │ (net/http)   │ │ (YAML → domain)    │  (outer layer)
   └─────────────────┘ └──────────────┘ └──────────────┘ └────────────────────┘
                              │ tui.Portfolio  │ web.Portfolio    │ implements
                              ▼                ▼                  ▼ usecase.Repository
                        ┌──────────────────────────────────────────────┐
                        │  usecase  (PortfolioService)                 │  application logic
                        └──────────────────────────────────────────────┘
                                              │
                                              ▼
                        ┌──────────────────────────────────────────────┐
                        │  domain  (entities + rules)                  │  core: no dependencies
                        └──────────────────────────────────────────────┘
```

**The dependency rule:** a package may import only packages *below* it in this
diagram. `domain` imports nothing from the project, and `usecase` imports only
`domain`. Adapters import `usecase` and/or `domain`, but never each other. Only
`cmd` knows about everything.

## The layers

### `internal/domain`: entities and business rules

Plain Go structs (`Portfolio`, `Experience`, `Project`, `Period`…) plus the rules
that must always hold, in `Portfolio.Validate()`: name and address are required,
slugs are unique, and a period can't end before it starts.

It has no tags, no YAML and no UI code. If you deleted every other folder, this
one would still compile.

### `internal/usecase`: application logic

`PortfolioService` is what the app *does* with the domain:

- loads content through a `Repository`, validates it, and fails fast at startup;
- sorts experiences and projects newest first;
- computes `YearsOfExperience()` from the dates, so "4+ years" never goes stale;
- derives the "SSH" contact from the site address;
- returns copies of slices, so one SSH session can't corrupt data another session is reading.

It defines the **`Repository` interface (an output port)**: "something that can
load a Portfolio". It doesn't know or care that the implementation reads YAML.

Time comes in through an injected `Clock`, so tests can freeze it.

### `internal/repository/yamlrepo`: storage adapter

Implements `usecase.Repository`. It decodes YAML into **DTOs** (structs with
`yaml:` tags) and maps them to domain types, parsing `YYYY-MM` dates into
`domain.Period`. YAML details stay here, so changing the file format never
touches the domain.

It uses `KnownFields(true)`, so a typo such as `nmae:` is an error rather than a
silently empty field.

### `internal/delivery/tui`: UI adapter

The Bubble Tea program. It depends on a **`tui.Portfolio` interface defined in
the tui package itself** (Go's "accept interfaces where they're consumed" idiom).
`usecase.PortfolioService` satisfies it implicitly, and tests pass a fake.

| File | Responsibility |
|---|---|
| `model.go` | state, `Update`, key handling, layout math |
| `view.go` | window chrome: address bar, tabs, footer, scrollbar |
| `pages.go` | one pure render function per page |
| `text.go` | wrapping, bullets, pills, gradient, truncation |
| `styles.go` | colors and Lip Gloss styles |

### `internal/delivery/sshserver`: transport adapter

Configures Wish: host key, open auth (it's a public site), idle timeout, and the
middleware chain `logging → activeterm → bubbletea`. It takes a `ModelFactory`, so
it has no idea it's serving a portfolio. It could serve any TUI.

### `internal/delivery/web`: HTTP adapter (landing page)

Serves the same domain over HTTP (Caddy adds HTTPS in front) so people who open
`rivaldo.dev` in a browser learn how to connect. One URL, two answers, picked by
the `Accept` header:

| Client | Gets |
|---|---|
| Browser (`Accept: text/html`) | `templates/index.html`: hero with a copy-able `ssh` command, a clickable replica of the TUI, connect steps, the host key fingerprint, controls, FAQ |
| `curl`, `wget`… | `terminal.go`: a Lip Gloss card with the command |

Like the TUI, it depends on its own small `web.Portfolio` interface, which the
same `PortfolioService` satisfies. That's why the browser preview and the SSH
app can never disagree: they render the same use-case output.

The page needs two values that aren't content: the command to show and the SSH
fingerprint. `cmd` passes both in `web.Config`, taking the fingerprint from
`sshserver.Server.Fingerprint()`. The adapters never talk to each other directly.

### `cmd/ssh-portfolio`: composition root

The only place where concrete types meet:

```go
repo := yamlrepo.New(content.Default)                        // pick a storage adapter
svc, _ := usecase.NewPortfolioService(ctx, repo, time.Now)   // build the use case
srv, _ := sshserver.New(cfg, func(ssh.Session) tea.Model {   // pick a delivery adapter
    return tui.New(svc)
})
web := web.New(web.Config{SSHCommand: "ssh rivaldo.dev",       // second delivery adapter
    Fingerprint: srv.Fingerprint()}, svc)                        // same use case
// run both with errgroup: if one fails, both shut down
```

`-local` swaps the SSH adapter for a local Bubble Tea program. Nothing else changes,
which shows the layers really are independent.

## Lifecycle of a visit

```
startup:  main → yamlrepo.Load → domain.Validate → PortfolioService (sorted, read-only)

web:      browser ──HTTPS:443──▶ Caddy ──HTTP:8080──▶ web.handleIndex ──▶ index.html
          curl    ──HTTPS:443──▶ Caddy ──HTTP:8080──▶ web.handleIndex ──▶ terminal card

visit:    ssh client ──TCP:22──▶ Wish
             logging      log connect / disconnect
             activeterm   reject sessions without a PTY
             bubbletea    ModelFactory → tui.New(svc)  (fresh state per visitor)
                 ▼
          Bubble Tea loop:  key / resize / wheel ──▶ Update ──▶ View ──▶ terminal
```

Content is loaded **once** and shared read-only. Each visitor gets their own
`tui.model`, so selections and scroll positions are never shared.

## Why bother, for a portfolio?

- **Content edits are safe.** Text lives in YAML, validated at startup and in tests.
  A broken edit fails `make test` instead of breaking the live site.
- **Each layer is testable on its own.** Domain rules, use-case logic, YAML parsing
  and UI navigation each have tests that need no network or SSH (`make test`).
- **Easy to extend.** The web landing page was added as one new adapter, with no
  changes to the domain or use cases. Content from Notion or a database would be
  the same: add an adapter. See [DEVELOPMENT.md](DEVELOPMENT.md).

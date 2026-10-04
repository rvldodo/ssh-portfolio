# Development

```sh
make local    # fastest loop: UI in this terminal
make run      # real SSH (ssh -p 23234 localhost) + landing page (http://localhost:8080)
make test     # go vet + tests for every layer
```

## Tests

| Package | What's tested | How |
|---|---|---|
| `domain` | `Period.String`, `Validate` reports every problem | plain values |
| `usecase` | sorting, years of experience, copies, error paths | `fakeRepo` + frozen `Clock` |
| `repository/yamlrepo` | embedded content is valid, date parsing, typo detection | YAML strings |
| `delivery/tui` | navigation, detail view, tab wrap-around, fits width | `fakePortfolio`, sending key messages to `Update` |
| `delivery/web` | browser gets HTML (command, fingerprint, escaping), curl gets the card, 404s, contact links | `httptest` + `fakePortfolio` |

UI tests drive the model directly, without a real terminal:

```go
m, _ := New(fakePortfolio{}).Update(tea.WindowSizeMsg{Width: 120, Height: 40})
m = press(m, char('2'), char('j'), keyEnter)   // go to Experience, select 2nd, open
strings.Contains(m.View().Content, "...")
```

## Recipes

### Add a field (e.g. a project URL)

Work from the inside out:

1. `domain/portfolio.go`: add `URL string` to `Project`.
2. `repository/yamlrepo`: add `URL string \`yaml:"url"\`` to `projectDTO` and map it in `toDomain`.
3. `delivery/tui/pages.go`: render it in `projectDetail`.
4. `content/portfolio.yaml`: fill it in. Document it in `docs/CONTENT.md`.

### Add a page (e.g. "Blog")

1. Add the data: entity in `domain`, getter on `PortfolioService`, YAML in the repository (as above).
2. Add the getter to the `tui.Portfolio` interface (and to `fakePortfolio` in the tests).
3. In `tui/model.go`: add a `blogPage` constant before `numPages`, plus entries in `pageNames` and `pagePaths`.
4. Add a case in `refresh()` that calls your new render function in `pages.go`.
5. Add the number key in `handleKey` (`"1"…"6"`).

### Load content from somewhere else (database, Notion, an API…)

Write a new adapter that implements `usecase.Repository`:

```go
type Repository interface {
    Load(ctx context.Context) (domain.Portfolio, error)
}
```

Then swap it in `cmd/ssh-portfolio/main.go`. The use case, domain and UI don't change.

### Add another way to view it (e.g. an HTTP page)

Create `internal/delivery/web` that takes the same `PortfolioService` and renders
HTML. Start it from `main.go` alongside the SSH server.

## Gotchas learned the hard way

- **`" " + multiLineString` only prefixes the first line.** Use
  `lipgloss.NewStyle().PaddingLeft(1).Render(s)` to pad every row.
- **Word-wrapping can leave trailing spaces** that make a line one column too wide.
  `trimLines` strips them.
- **Each server generates its own host key.** Running different servers on the same
  `localhost` port triggers SSH's "host identification has changed" warning. Use
  different ports, or `ssh-keygen -R "[localhost]:23234"`.
- **Bubble Tea v2 declares terminal features in `View()`** (`AltScreen`, `MouseMode`,
  `WindowTitle`, colors), not as program options.

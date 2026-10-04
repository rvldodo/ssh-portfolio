# Content

Everything visitors see comes from [`content/portfolio.yaml`](../content/portfolio.yaml).
It's embedded into the binary, so **rebuild or redeploy after editing**. To try
changes without rebuilding, run:

```sh
go run ./cmd/ssh-portfolio -local -content content/portfolio.yaml
```

Run `make test` after editing. It fails if the YAML has a typo, a bad date or a
duplicate slug.

## Schema

### `site`

| Field | Example | Notes |
|---|---|---|
| `address` | `portfolio.rvldodo.cloud` | Required. Shown in the address bar and as the `ssh …` contact. |

### `profile`

| Field | Notes |
|---|---|
| `name` | Required. Rendered with a gradient on Home. |
| `title`, `location` | Subtitle line. Years of experience is appended automatically. |
| `summary` | Paragraph on Home. Use `>` for multi-line text. |
| `now` | "Now" section on Home (optional). |
| `highlights` | List of `{ label, detail }` for the "At a glance" table. |

### `experiences` (list, shown newest first)

| Field | Required | Example |
|---|---|---|
| `slug` | ✓ unique | `woori` |
| `company`, `role` | ✓ | `Waterhub`, `Backend Developer` |
| `org` | | `Startup` |
| `employment` | | `Freelance` |
| `location` | | `Remote, Indonesia` |
| `start` | ✓ | `2024-05` (YYYY-MM) |
| `end` | | `2024-07`; leave it out, or write `present`, for current roles |
| `about` | | Short company description |
| `achievements` | | List of bullet strings |
| `stack` | | List of tags, shown as pills |

### `projects` (list, shown newest first)

Same as experiences, but uses `name` and `kind` instead of company, role and
employment, and `highlights` instead of `achievements`. An optional `url`
(e.g. `gerakanturuntangan.com`) is shown under the dates.

### `skills`

List of `{ group, skills: [...] }`. Shown in the order written.

### `education`

List of `{ school, field, location, year }`. Shown newest first.

### `contacts`

List of `{ label, value }`. An `SSH` entry is added automatically from `site.address`.

## YAML gotchas

- **Quote strings that contain `: `** (colon + space), or YAML reads them as a map:
  `- "Built X: request signing, retries"`.
- **Quote values with commas inside `{ }` flow maps:** `{ location: "Kazan, Russia" }`.
- Unknown field names are rejected, so a typo like `acheivements:` fails loudly.

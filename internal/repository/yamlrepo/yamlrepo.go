// Package yamlrepo implements usecase.Repository by reading portfolio
// content from YAML.
//
// YAML is a storage detail, so it stays in this adapter: the yaml-tagged
// structs below are data transfer objects (DTOs) that get mapped into clean
// domain types. Swapping YAML for JSON, a database, or a CMS means writing
// another adapter like this one; nothing else changes.
package yamlrepo

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"

	"ssh-portfolio/internal/domain"
)

// Repository reads a portfolio from YAML bytes.
type Repository struct {
	data []byte
}

// New returns a repository over in-memory YAML, e.g. the embedded default
// content from the content package.
func New(data []byte) *Repository { return &Repository{data: data} }

// FromFile returns a repository over a YAML file on disk.
func FromFile(path string) (*Repository, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return New(data), nil
}

// Load parses the YAML and maps it to domain types.
func (r *Repository) Load(context.Context) (domain.Portfolio, error) {
	var doc document
	dec := yaml.NewDecoder(bytes.NewReader(r.data))
	dec.KnownFields(true) // a typo in a field name is an error, not silently ignored
	if err := dec.Decode(&doc); err != nil {
		return domain.Portfolio{}, fmt.Errorf("parse yaml: %w", err)
	}
	return doc.toDomain()
}

// ---------------------------------------------------------------------------
// DTOs: the on-disk shape. Documented in docs/CONTENT.md.

type document struct {
	Site struct {
		Address string `yaml:"address"`
	} `yaml:"site"`
	Profile struct {
		Name       string `yaml:"name"`
		Title      string `yaml:"title"`
		Location   string `yaml:"location"`
		Summary    string `yaml:"summary"`
		Now        string `yaml:"now"`
		Highlights []struct {
			Label  string `yaml:"label"`
			Detail string `yaml:"detail"`
		} `yaml:"highlights"`
	} `yaml:"profile"`
	Experiences []experienceDTO `yaml:"experiences"`
	Projects    []projectDTO    `yaml:"projects"`
	Skills      []struct {
		Group  string   `yaml:"group"`
		Skills []string `yaml:"skills"`
	} `yaml:"skills"`
	Education []struct {
		School   string `yaml:"school"`
		Field    string `yaml:"field"`
		Location string `yaml:"location"`
		Year     int    `yaml:"year"`
	} `yaml:"education"`
	Contacts []struct {
		Label string `yaml:"label"`
		Value string `yaml:"value"`
	} `yaml:"contacts"`
}

type experienceDTO struct {
	Slug         string   `yaml:"slug"`
	Company      string   `yaml:"company"`
	Org          string   `yaml:"org"`
	Role         string   `yaml:"role"`
	Employment   string   `yaml:"employment"`
	Location     string   `yaml:"location"`
	Start        string   `yaml:"start"`
	End          string   `yaml:"end"`
	About        string   `yaml:"about"`
	Achievements []string `yaml:"achievements"`
	Stack        []string `yaml:"stack"`
}

type projectDTO struct {
	Slug       string   `yaml:"slug"`
	Name       string   `yaml:"name"`
	Kind       string   `yaml:"kind"`
	Location   string   `yaml:"location"`
	Start      string   `yaml:"start"`
	End        string   `yaml:"end"`
	About      string   `yaml:"about"`
	Highlights []string `yaml:"highlights"`
	Stack      []string `yaml:"stack"`
}

func (d document) toDomain() (domain.Portfolio, error) {
	p := domain.Portfolio{
		Site: domain.Site{Address: d.Site.Address},
		Profile: domain.Profile{
			Name:     d.Profile.Name,
			Title:    d.Profile.Title,
			Location: d.Profile.Location,
			Summary:  strings.TrimSpace(d.Profile.Summary),
			Now:      strings.TrimSpace(d.Profile.Now),
		},
	}
	for _, h := range d.Profile.Highlights {
		p.Profile.Highlights = append(
			p.Profile.Highlights,
			domain.Highlight{Label: h.Label, Detail: h.Detail},
		)
	}

	for _, e := range d.Experiences {
		period, err := parsePeriod(e.Start, e.End)
		if err != nil {
			return domain.Portfolio{}, fmt.Errorf("experience %q: %w", e.Slug, err)
		}
		p.Experiences = append(p.Experiences, domain.Experience{
			Slug: e.Slug, Company: e.Company, Org: e.Org, Role: e.Role,
			Employment: e.Employment, Location: e.Location, Period: period,
			About: strings.TrimSpace(e.About), Achievements: e.Achievements, Stack: e.Stack,
		})
	}

	for _, pr := range d.Projects {
		period, err := parsePeriod(pr.Start, pr.End)
		if err != nil {
			return domain.Portfolio{}, fmt.Errorf("project %q: %w", pr.Slug, err)
		}
		p.Projects = append(p.Projects, domain.Project{
			Slug: pr.Slug, Name: pr.Name, Kind: pr.Kind, Location: pr.Location, Period: period,
			About: strings.TrimSpace(pr.About), Highlights: pr.Highlights, Stack: pr.Stack,
		})
	}

	for _, s := range d.Skills {
		p.Skills = append(p.Skills, domain.SkillGroup{Name: s.Group, Skills: s.Skills})
	}
	for _, e := range d.Education {
		p.Education = append(
			p.Education,
			domain.Education{School: e.School, Field: e.Field, Location: e.Location, Year: e.Year},
		)
	}
	for _, c := range d.Contacts {
		p.Contacts = append(p.Contacts, domain.Contact{Label: c.Label, Value: c.Value})
	}
	return p, nil
}

// parsePeriod reads "YYYY-MM" dates. An empty end or "present" means ongoing.
func parsePeriod(start, end string) (domain.Period, error) {
	s, err := time.Parse("2006-01", start)
	if err != nil {
		return domain.Period{}, fmt.Errorf("start %q: want YYYY-MM", start)
	}
	period := domain.Period{Start: s}
	if end == "" || strings.EqualFold(end, "present") {
		return period, nil
	}
	e, err := time.Parse("2006-01", end)
	if err != nil {
		return domain.Period{}, fmt.Errorf("end %q: want YYYY-MM or present", end)
	}
	period.End = &e
	return period, nil
}

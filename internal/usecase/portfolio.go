// Package usecase holds the application logic: what the portfolio can do,
// independent of how it is stored or displayed.
//
// It depends only on the domain. Storage is reached through the Repository
// interface (an "output port") that this package defines and outer layers
// implement, so the dependency arrow always points inward.
package usecase

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"time"

	"ssh-portfolio/internal/domain"
)

// Repository loads portfolio content from some source (a YAML file, a
// database, an API...). Implementations live in internal/repository.
type Repository interface {
	Load(ctx context.Context) (domain.Portfolio, error)
}

// Clock returns the current time. Injected so tests can freeze time.
type Clock func() time.Time

// PortfolioService serves portfolio content to delivery adapters such as
// the TUI. Content is loaded and validated once at startup, then shared
// read-only by every session.
type PortfolioService struct {
	p   domain.Portfolio
	now Clock
}

// NewPortfolioService loads content from repo, validates it, and prepares
// it for display (newest entries first).
func NewPortfolioService(
	ctx context.Context,
	repo Repository,
	now Clock,
) (*PortfolioService, error) {
	p, err := repo.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("load portfolio: %w", err)
	}
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("invalid portfolio:\n%w", err)
	}

	// Newest first. Stable sort keeps the authored order for equal dates.
	slices.SortStableFunc(p.Experiences, func(a, b domain.Experience) int {
		return b.Period.Start.Compare(a.Period.Start)
	})
	slices.SortStableFunc(p.Projects, func(a, b domain.Project) int {
		return b.Period.Start.Compare(a.Period.Start)
	})
	slices.SortStableFunc(p.Education, func(a, b domain.Education) int {
		return cmp.Compare(b.Year, a.Year)
	})

	return &PortfolioService{p: p, now: now}, nil
}

// The getters return copies of slices so callers can't mutate shared state
// that other SSH sessions are reading.

func (s *PortfolioService) Site() domain.Site       { return s.p.Site }
func (s *PortfolioService) Profile() domain.Profile { return s.p.Profile }

func (s *PortfolioService) Experiences() []domain.Experience { return slices.Clone(s.p.Experiences) }
func (s *PortfolioService) Projects() []domain.Project       { return slices.Clone(s.p.Projects) }
func (s *PortfolioService) Skills() []domain.SkillGroup      { return slices.Clone(s.p.Skills) }
func (s *PortfolioService) Education() []domain.Education    { return slices.Clone(s.p.Education) }

// Contacts returns the configured contacts plus an "SSH" entry derived from
// the site address, so the address only has to be written once.
func (s *PortfolioService) Contacts() []domain.Contact {
	return append(
		slices.Clone(s.p.Contacts),
		domain.Contact{Label: "SSH", Value: "ssh " + s.p.Site.Address},
	)
}

// YearsOfExperience counts whole years since the earliest experience began.
// It is computed rather than written down so it never goes stale.
func (s *PortfolioService) YearsOfExperience() int {
	if len(s.p.Experiences) == 0 {
		return 0
	}
	first := s.p.Experiences[0].Period.Start
	for _, e := range s.p.Experiences[1:] {
		if e.Period.Start.Before(first) {
			first = e.Period.Start
		}
	}
	now := s.now()
	years := now.Year() - first.Year()
	if now.Month() < first.Month() {
		years--
	}
	return max(0, years)
}

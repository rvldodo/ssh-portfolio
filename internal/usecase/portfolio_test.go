package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"ssh-portfolio/internal/domain"
)

// fakeRepo is an in-memory Repository. Because the use case depends on an
// interface, tests need no files, YAML, or network.
type fakeRepo struct {
	p   domain.Portfolio
	err error
}

func (f fakeRepo) Load(context.Context) (domain.Portfolio, error) { return f.p, f.err }

func month(s string) time.Time {
	t, _ := time.Parse("2006-01", s)
	return t
}

func fixedClock(s string) Clock { return func() time.Time { return month(s) } }

func samplePortfolio() domain.Portfolio {
	exp := func(slug, start string) domain.Experience {
		return domain.Experience{Slug: slug, Company: "C", Role: "R", Period: domain.Period{Start: month(start)}}
	}
	return domain.Portfolio{
		Site:        domain.Site{Address: "example.dev"},
		Profile:     domain.Profile{Name: "Ada"},
		Experiences: []domain.Experience{exp("old", "2022-09"), exp("new", "2024-09"), exp("mid", "2023-04")},
		Contacts:    []domain.Contact{{Label: "Email", Value: "ada@example.dev"}},
	}
}

func TestExperiencesSortedNewestFirst(t *testing.T) {
	svc, err := NewPortfolioService(context.Background(), fakeRepo{p: samplePortfolio()}, fixedClock("2026-10"))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, e := range svc.Experiences() {
		got = append(got, e.Slug)
	}
	if want := []string{"new", "mid", "old"}; !equal(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}

func TestYearsOfExperience(t *testing.T) {
	tests := []struct {
		now  string
		want int
	}{
		{"2026-10", 4}, // Sep 2022 -> Oct 2026
		{"2026-08", 3}, // one month short of 4 years
		{"2022-09", 0},
	}
	for _, tt := range tests {
		svc, err := NewPortfolioService(context.Background(), fakeRepo{p: samplePortfolio()}, fixedClock(tt.now))
		if err != nil {
			t.Fatal(err)
		}
		if got := svc.YearsOfExperience(); got != tt.want {
			t.Errorf("now=%s: YearsOfExperience() = %d, want %d", tt.now, got, tt.want)
		}
	}
}

func TestContactsIncludeSSHAddress(t *testing.T) {
	svc, _ := NewPortfolioService(context.Background(), fakeRepo{p: samplePortfolio()}, fixedClock("2026-10"))
	c := svc.Contacts()
	if last := c[len(c)-1]; last.Label != "SSH" || last.Value != "ssh example.dev" {
		t.Errorf("last contact = %+v, want SSH entry", last)
	}
}

func TestGettersReturnCopies(t *testing.T) {
	svc, _ := NewPortfolioService(context.Background(), fakeRepo{p: samplePortfolio()}, fixedClock("2026-10"))
	svc.Experiences()[0].Slug = "mutated"
	if svc.Experiences()[0].Slug == "mutated" {
		t.Error("caller mutated shared service state")
	}
}

func TestNewPortfolioServiceErrors(t *testing.T) {
	if _, err := NewPortfolioService(context.Background(), fakeRepo{err: errors.New("disk on fire")}, time.Now); err == nil {
		t.Error("want error when repository fails")
	}
	bad := samplePortfolio()
	bad.Profile.Name = ""
	if _, err := NewPortfolioService(context.Background(), fakeRepo{p: bad}, time.Now); err == nil {
		t.Error("want error for invalid portfolio")
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

package domain

import (
	"strings"
	"testing"
	"time"
)

func month(s string) time.Time {
	t, err := time.Parse("2006-01", s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestPeriodString(t *testing.T) {
	end := month("2023-08")
	tests := []struct {
		p    Period
		want string
	}{
		{Period{Start: month("2023-04"), End: &end}, "Apr 2023 – Aug 2023"},
		{Period{Start: month("2024-09")}, "Sep 2024 – Present"},
	}
	for _, tt := range tests {
		if got := tt.p.String(); got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	}
}

func validPortfolio() Portfolio {
	return Portfolio{
		Site:    Site{Address: "example.dev"},
		Profile: Profile{Name: "Ada"},
		Experiences: []Experience{
			{Slug: "a", Company: "Acme", Role: "Dev", Period: Period{Start: month("2024-01")}},
		},
		Projects: []Project{{Slug: "p", Name: "Proj", Period: Period{Start: month("2024-01")}}},
	}
}

func TestValidateAcceptsValidPortfolio(t *testing.T) {
	if err := validPortfolio().Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil", err)
	}
}

func TestValidateReportsEveryProblem(t *testing.T) {
	p := validPortfolio()
	p.Profile.Name = ""
	p.Experiences = append(
		p.Experiences,
		Experience{Slug: "a", Company: "Dup", Role: "Dev", Period: Period{Start: month("2024-01")}},
	)
	before := month("2023-01")
	p.Projects[0].Period.End = &before

	err := p.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want errors")
	}
	for _, want := range []string{"name is required", `duplicate slug "a"`, "ends before it starts"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

package yamlrepo

import (
	"context"
	"strings"
	"testing"

	"ssh-portfolio/content"
)

// The shipped content must always parse and pass validation.
func TestEmbeddedContentIsValid(t *testing.T) {
	p, err := New(content.Default).Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("embedded content is invalid:\n%v", err)
	}
	if len(p.Experiences) == 0 || len(p.Projects) == 0 {
		t.Error("expected experiences and projects")
	}
}

func TestParsesPeriods(t *testing.T) {
	yml := `
experiences:
  - { slug: a, start: 2023-04, end: 2023-08 }
  - { slug: b, start: 2024-09 }
  - { slug: c, start: 2024-01, end: Present }
`
	p, err := New([]byte(yml)).Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got := p.Experiences[0].Period.String(); got != "Apr 2023 – Aug 2023" {
		t.Errorf("a = %q", got)
	}
	if !p.Experiences[1].Period.Ongoing() || !p.Experiences[2].Period.Ongoing() {
		t.Error("missing end and 'Present' should both mean ongoing")
	}
}

func TestRejectsBadInput(t *testing.T) {
	tests := map[string]struct{ yml, want string }{
		"bad date":      {"experiences: [{ slug: a, start: September 2024 }]", "want YYYY-MM"},
		"unknown field": {"profile: { nmae: typo }", "nmae"},
	}
	for name, tt := range tests {
		_, err := New([]byte(tt.yml)).Load(context.Background())
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: err = %v, want mention of %q", name, err, tt.want)
		}
	}
}

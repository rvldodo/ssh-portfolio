package tui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"ssh-portfolio/internal/domain"
)

// fakePortfolio implements Portfolio with fixed data, so UI tests don't
// depend on the real content or the use case layer.
type fakePortfolio struct{}

func (fakePortfolio) Site() domain.Site { return domain.Site{Address: "example.dev"} }
func (fakePortfolio) Profile() domain.Profile {
	return domain.Profile{Name: "Ada Lovelace", Title: "Engineer"}
}
func (fakePortfolio) Experiences() []domain.Experience {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	return []domain.Experience{
		{
			Slug:         "first",
			Company:      "Acme",
			Role:         "Engineer",
			Period:       domain.Period{Start: start},
			Achievements: []string{"Built the first thing"},
		},
		{
			Slug:         "second",
			Company:      "Globex",
			Role:         "Lead",
			Period:       domain.Period{Start: start},
			Achievements: []string{"Led the second thing"},
		},
	}
}
func (fakePortfolio) Projects() []domain.Project    { return nil }
func (fakePortfolio) Skills() []domain.SkillGroup   { return nil }
func (fakePortfolio) Education() []domain.Education { return nil }
func (fakePortfolio) Contacts() []domain.Contact    { return nil }
func (fakePortfolio) YearsOfExperience() int        { return 3 }

func press(m tea.Model, keys ...tea.KeyPressMsg) tea.Model {
	for _, k := range keys {
		m, _ = m.Update(k)
	}
	return m
}

var (
	keyEnter = tea.KeyPressMsg{Code: tea.KeyEnter}
	keyEsc   = tea.KeyPressMsg{Code: tea.KeyEscape}
	keyRight = tea.KeyPressMsg{Code: tea.KeyRight}
)

func char(r rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: r, Text: string(r)} }

func newTestModel() tea.Model {
	m, _ := New(fakePortfolio{}).Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return m
}

func TestNavigateToExperienceDetailAndBack(t *testing.T) {
	m := press(newTestModel(), char('2'), char('j'), keyEnter)
	mm := m.(model)
	if mm.page != experiencePage || mm.cursor[experiencePage] != 1 || !mm.open[experiencePage] {
		t.Fatalf(
			"state = page %d cursor %d open %v",
			mm.page,
			mm.cursor[experiencePage],
			mm.open[experiencePage],
		)
	}
	if !strings.Contains(m.View().Content, "Led the second thing") {
		t.Error("detail view should show the selected experience")
	}

	mm = press(m, keyEsc).(model)
	if mm.open[experiencePage] {
		t.Error("esc should close the detail view")
	}
}

func TestTabsWrapAround(t *testing.T) {
	m := newTestModel()
	for range numPages {
		m = press(m, keyRight)
	}
	if p := m.(model).page; p != homePage {
		t.Errorf("after %d rights, page = %d, want home", numPages, p)
	}
}

func TestHomeShowsComputedYears(t *testing.T) {
	if v := newTestModel().View().Content; !strings.Contains(v, "3+ years") {
		t.Error("home should show years of experience from the Portfolio")
	}
}

func TestViewFitsTerminalWidth(t *testing.T) {
	for _, w := range []int{60, 80, 120} {
		m, _ := New(fakePortfolio{}).Update(tea.WindowSizeMsg{Width: w, Height: 30})
		for i, line := range strings.Split(m.View().Content, "\n") {
			if got := len([]rune(stripANSI(line))); got > w {
				t.Errorf("width %d: line %d is %d columns", w, i, got)
			}
		}
	}
}

// stripANSI removes escape sequences so we can count visible characters.
func stripANSI(s string) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			in = true
		case in && (r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z'):
			in = false
		case !in:
			b.WriteRune(r)
		}
	}
	return b.String()
}

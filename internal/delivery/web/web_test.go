package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ssh-portfolio/internal/domain"
)

type fakePortfolio struct{}

func (fakePortfolio) Site() domain.Site { return domain.Site{Address: "example.dev"} }
func (fakePortfolio) Profile() domain.Profile {
	return domain.Profile{Name: "Ada Lovelace", Title: "Engineer", Location: "London"}
}
func (fakePortfolio) Experiences() []domain.Experience {
	return []domain.Experience{{
		Slug: "acme", Company: "Acme", Role: "Engineer",
		Period:       domain.Period{Start: time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)},
		Achievements: []string{"Built <the> engine"},
	}}
}
func (fakePortfolio) Projects() []domain.Project  { return nil }
func (fakePortfolio) Skills() []domain.SkillGroup { return nil }
func (fakePortfolio) Contacts() []domain.Contact {
	return []domain.Contact{{Label: "Email", Value: "ada@example.dev"}, {Label: "SSH", Value: "ssh example.dev"}}
}
func (fakePortfolio) YearsOfExperience() int { return 2 }

func get(t *testing.T, path, accept string) (*http.Response, string) {
	t.Helper()
	srv := New(Config{SSHCommand: "ssh example.dev", Fingerprint: "SHA256:abc123"}, fakePortfolio{})
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	body, _ := io.ReadAll(rec.Result().Body)
	return rec.Result(), string(body)
}

func TestBrowserGetsLandingPage(t *testing.T) {
	res, body := get(t, "/", "text/html,application/xhtml+xml,*/*;q=0.8")
	if res.StatusCode != 200 || !strings.HasPrefix(res.Header.Get("Content-Type"), "text/html") {
		t.Fatalf("status %d, type %q", res.StatusCode, res.Header.Get("Content-Type"))
	}
	for _, want := range []string{
		"ssh example.dev",               // the command to copy
		"SHA256:abc123",                 // fingerprint to verify
		"Ada Lovelace",                  // content from the Portfolio
		"Sep 2024 – Present",            // domain.Period rendered via String()
		`href="mailto:ada@example.dev"`, // contacts become links
		"Built &lt;the&gt; engine",      // html/template escapes content
	} {
		if !strings.Contains(body, want) {
			t.Errorf("page is missing %q", want)
		}
	}
}

func TestCurlGetsTerminalCard(t *testing.T) {
	res, body := get(t, "/", "*/*")
	if !strings.HasPrefix(res.Header.Get("Content-Type"), "text/plain") {
		t.Fatalf("type %q, want text/plain", res.Header.Get("Content-Type"))
	}
	if !strings.Contains(body, "ssh example.dev") || strings.Contains(body, "<html") {
		t.Errorf("terminal card should be plain text with the command:\n%s", body)
	}
	if !strings.Contains(res.Header.Get("Vary"), "Accept") {
		t.Error("response varies by Accept, so caches must be told")
	}
}

func TestHealthz(t *testing.T) {
	if res, _ := get(t, "/healthz", ""); res.StatusCode != 200 {
		t.Errorf("healthz status %d", res.StatusCode)
	}
}

func TestUnknownPathIs404(t *testing.T) {
	if res, _ := get(t, "/nope", "text/html"); res.StatusCode != 404 {
		t.Errorf("status %d, want 404", res.StatusCode)
	}
}

func TestContactURL(t *testing.T) {
	tests := map[string]string{
		"me@example.com":             "mailto:me@example.com",
		"linkedin.com/in/someone":    "https://linkedin.com/in/someone",
		"https://github.com/someone": "https://github.com/someone",
		"ssh example.dev":            "",
	}
	for in, want := range tests {
		if got := contactURL(domain.Contact{Value: in}); got != want {
			t.Errorf("contactURL(%q) = %q, want %q", in, got, want)
		}
	}
}

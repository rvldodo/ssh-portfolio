package web

import (
	"bytes"
	"embed"
	"html/template"
	"net/http"
	"strings"

	"charm.land/log/v2"

	"ssh-portfolio/internal/domain"
)

//go:embed templates/index.html
var templateFS embed.FS

var indexTmpl = template.Must(template.New("index.html").Funcs(template.FuncMap{
	"contactURL": contactURL,
}).ParseFS(templateFS, "templates/index.html"))

// pageData is the view model for the landing page template.
type pageData struct {
	Profile     domain.Profile
	Years       int
	Address     string
	Command     string
	Fingerprint string
	Experiences []domain.Experience
	Projects    []domain.Project
	Skills      []domain.SkillGroup
	Contacts    []domain.Contact
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	// The same URL answers browsers and terminals differently, so caches
	// must key on these headers.
	w.Header().Set("Vary", "Accept, User-Agent")
	w.Header().Set("Cache-Control", "public, max-age=300")

	if !wantsHTML(r) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(terminalCard(s.data, s.cfg.SSHCommand)))
		return
	}

	data := pageData{
		Profile:     s.data.Profile(),
		Years:       s.data.YearsOfExperience(),
		Address:     s.data.Site().Address,
		Command:     s.cfg.SSHCommand,
		Fingerprint: s.cfg.Fingerprint,
		Experiences: s.data.Experiences(),
		Projects:    s.data.Projects(),
		Skills:      s.data.Skills(),
		Contacts:    s.data.Contacts(),
	}

	// Render to a buffer first so a template error becomes a clean 500
	// instead of a half-written page.
	var buf bytes.Buffer
	if err := indexTmpl.Execute(&buf, data); err != nil {
		log.Error("render index", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

// wantsHTML reports whether the client is a browser. Browsers always list
// text/html in Accept; curl, wget, and HTTPie send */* or nothing.
func wantsHTML(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

// contactURL turns a contact value into a link, or "" if it isn't one.
func contactURL(c domain.Contact) string {
	v := c.Value
	switch {
	case strings.HasPrefix(v, "ssh "):
		return ""
	case strings.Contains(v, "@") && !strings.Contains(v, "/"):
		return "mailto:" + v
	case strings.HasPrefix(v, "http://"), strings.HasPrefix(v, "https://"):
		return v
	case strings.Contains(v, "."):
		return "https://" + v
	}
	return ""
}

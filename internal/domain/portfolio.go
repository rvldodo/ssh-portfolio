// Package domain holds the core business entities of the portfolio.
//
// This is the innermost layer: it depends on nothing but the standard
// library. It knows nothing about YAML, SSH, terminals, or Bubble Tea, so
// the rest of the system can change without touching it.
package domain

// Portfolio is the aggregate root: everything a visitor can see.
type Portfolio struct {
	Site        Site
	Profile     Profile
	Experiences []Experience
	Projects    []Project
	Skills      []SkillGroup
	Education   []Education
	Contacts    []Contact
}

// Site describes where the portfolio is published.
type Site struct {
	// Address is the host people type after `ssh`, e.g. "rivaldo.dev".
	Address string
}

// Profile is the person the portfolio is about.
type Profile struct {
	Name       string
	Title      string
	Location   string
	Summary    string
	Now        string // what they are working on currently
	Highlights []Highlight
}

// Highlight is a short "at a glance" fact, e.g. {"Docker Swarm", "zero-downtime deploys"}.
type Highlight struct {
	Label  string
	Detail string
}

// Experience is one job or engagement.
type Experience struct {
	Slug         string // stable identifier, unique among experiences
	Company      string
	Org          string // kind of organization, e.g. "Startup"
	Role         string
	Employment   string // e.g. "Full-time", "Freelance"
	Location     string
	Period       Period
	About        string
	Achievements []string
	Stack        []string
}

// Project is a piece of client or personal work.
type Project struct {
	Slug       string // stable identifier, unique among projects
	Name       string
	Kind       string // e.g. "Company profile with CMS"
	Location   string
	Period     Period
	About      string
	Highlights []string
	Stack      []string
}

// SkillGroup is a named set of skills, e.g. "Backend".
type SkillGroup struct {
	Name   string
	Skills []string
}

// Education is one school attended.
type Education struct {
	School   string
	Field    string
	Location string
	Year     int
}

// Contact is one way to reach the person, e.g. {"Email", "me@example.com"}.
type Contact struct {
	Label string
	Value string
}

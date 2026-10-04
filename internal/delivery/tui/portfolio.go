// Package tui is the terminal UI delivery adapter, built with Bubble Tea.
//
// It renders the portfolio like a website inside a "browser window" card:
//
//	● ● ●   ssh rivaldo.dev/experience          <- address bar (changes per page)
//	 Home   Experience   Projects   Skills  ... <- nav tabs
//	────────────────────────────────────────
//	page content (scrollable viewport)    ┃     <- scrollbar
//	footer key help
//
// Files:
//
//	portfolio.go  the Portfolio interface this package depends on
//	model.go      state and Update (keys, resizes, navigation)
//	view.go       the window chrome: address bar, tabs, footer, scrollbar
//	pages.go      one render function per page
//	text.go       small text helpers (wrapping, bullets, pills, gradient)
//	styles.go     colors and Lip Gloss styles
package tui

import "ssh-portfolio/internal/domain"

// Portfolio is everything the TUI needs to read. It's defined here, where
// it's consumed (a Go idiom), and implemented by
// usecase.PortfolioService. Tests can pass a fake instead.
type Portfolio interface {
	Site() domain.Site
	Profile() domain.Profile
	Experiences() []domain.Experience
	Projects() []domain.Project
	Skills() []domain.SkillGroup
	Education() []domain.Education
	Contacts() []domain.Contact
	YearsOfExperience() int
}

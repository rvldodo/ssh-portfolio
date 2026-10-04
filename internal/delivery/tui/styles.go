package tui

import "charm.land/lipgloss/v2"

const padX = 3 // horizontal padding inside the card border

// Colors. Lip Gloss downsamples automatically for terminals with fewer colors.
var (
	colBg      = lipgloss.Color("#0b0b0f")
	colFg      = lipgloss.Color("#a8a8b3")
	colBorder  = lipgloss.Color("236")
	colDim     = lipgloss.Color("243")
	colFaint   = lipgloss.Color("239")
	colBright  = lipgloss.Color("255")
	colAccent  = lipgloss.Color("#2dd4bf")
	colAccent2 = lipgloss.Color("#a78bfa")
)

var (
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colBorder).
			Padding(1, padX)
	h1Style     = lipgloss.NewStyle().Foreground(colBright).Bold(true)
	h2Style     = lipgloss.NewStyle().Foreground(colAccent).Bold(true)
	textStyle   = lipgloss.NewStyle().Foreground(colFg)
	dimStyle    = lipgloss.NewStyle().Foreground(colDim)
	faintStyle  = lipgloss.NewStyle().Foreground(colFaint)
	brightStyle = lipgloss.NewStyle().Foreground(colBright)
	accentStyle = lipgloss.NewStyle().Foreground(colAccent)
	selStyle    = lipgloss.NewStyle().Foreground(colAccent).Bold(true)
	pillStyle   = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)
	urlStyle = lipgloss.NewStyle().
			Foreground(colDim).
			Background(lipgloss.Color("234")).
			Padding(0, 2)
	tabStyle   = lipgloss.NewStyle().Foreground(colDim).Padding(0, 1)
	tabOnStyle = lipgloss.NewStyle().
			Foreground(colBg).
			Background(colAccent).
			Bold(true).
			Padding(0, 1)
	ruleStyle  = lipgloss.NewStyle().Foreground(colBorder)
	thumbStyle = lipgloss.NewStyle().Foreground(colDim)
)

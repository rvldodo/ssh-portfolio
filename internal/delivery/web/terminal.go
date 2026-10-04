package web

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// terminalCard is what `curl rivaldo.dev` prints: a small card in the same
// style as the TUI that points to the real thing.
func terminalCard(data Portfolio, command string) string {
	var (
		border = lipgloss.Color("236")
		dim    = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
		bright = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
		accent = lipgloss.NewStyle().Foreground(lipgloss.Color("#2dd4bf")).Bold(true)
		cmd    = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Background(lipgloss.Color("235")).Bold(true).Padding(0, 1)
	)

	p := data.Profile()
	dots := lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render("●") + " " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("●") + " " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("41")).Render("●")

	body := strings.Join([]string{
		dots,
		"",
		bright.Render(p.Name),
		dim.Render(p.Title + " · " + p.Location),
		"",
		"My portfolio lives in your terminal:",
		"",
		accent.Render("$ ") + cmd.Render(command),
		"",
		dim.Render("←/→ pages  •  ↑/↓ scroll  •  enter open  •  q quit"),
	}, "\n")

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(1, 3).
		Render(body)
	return "\n" + card + "\n\n"
}

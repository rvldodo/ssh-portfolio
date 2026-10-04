package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// View draws the browser-window card around the current page. In Bubble
// Tea v2 the View also declares terminal features: alt screen, mouse,
// window title, and default colors.
func (m model) View() tea.View {
	inner := m.cardWidth() - 2 - padX*2

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(m.textWidth()).MaxWidth(m.textWidth()).Render(m.vp.View()),
		lipgloss.NewStyle().PaddingLeft(1).Render(m.scrollbar()), // pads every row, unlike " "+s
	)

	card := cardStyle.Render(strings.Join([]string{
		m.addressBar(inner),
		"",
		m.tabs(),
		ruleStyle.Render(strings.Repeat("─", inner)),
		"",
		body,
		"",
		m.footer(inner),
	}, "\n"))

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.WindowTitle = m.data.Profile().Name + " — " + pageNames[m.page]
	v.BackgroundColor = colBg
	v.ForegroundColor = colFg
	return v
}

func (m model) addressBar(width int) string {
	dots := lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render("●") + " " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("●") + " " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("41")).Render("●")
	url := urlStyle.Render(
		faintStyle.Render(
			"ssh ",
		) + brightStyle.Render(
			m.data.Site().Address,
		) + dimStyle.Render(
			pagePaths[m.page],
		),
	)
	gap := strings.Repeat(" ", max(1, width-lipgloss.Width(dots)-lipgloss.Width(url)-4)/2)
	return dots + gap + "  " + url
}

func (m model) tabs() string {
	parts := make([]string, 0, numPages)
	for i, name := range pageNames {
		if page(i) == m.page {
			parts = append(parts, tabOnStyle.Render(name))
		} else {
			parts = append(parts, tabStyle.Render(name))
		}
	}
	return strings.Join(parts, " ")
}

func (m model) footer(width int) string {
	var help string
	switch {
	case m.open[m.page]:
		help = "↑/↓ scroll  •  esc back  •  ←/→ pages  •  q quit"
	case m.listLen() > 0:
		help = "↑/↓ select  •  enter open  •  ←/→ pages  •  q quit"
	default:
		help = "↑/↓ scroll  •  ←/→ or 1-5 pages  •  q quit"
	}
	pos := ""
	if m.vp.TotalLineCount() > m.vp.Height() {
		pos = fmt.Sprintf("%3.f%%", m.vp.ScrollPercent()*100)
	}
	gap := strings.Repeat(" ", max(1, width-lipgloss.Width(help)-lipgloss.Width(pos)))
	return dimStyle.Render(help) + gap + faintStyle.Render(pos)
}

// scrollbar draws a track (│) with a thumb (┃) whose size and position
// mirror the visible fraction and scroll offset. Blank when nothing scrolls.
func (m model) scrollbar() string {
	h, total := m.vp.Height(), m.vp.TotalLineCount()
	if total <= h {
		return strings.TrimSuffix(strings.Repeat(" \n", h), "\n")
	}
	thumbH := max(1, h*h/total)
	thumbTop := int(float64(h-thumbH) * m.vp.ScrollPercent())
	rows := make([]string, h)
	for i := range rows {
		rows[i] = ruleStyle.Render("│")
		if i >= thumbTop && i < thumbTop+thumbH {
			rows[i] = thumbStyle.Render("┃")
		}
	}
	return strings.Join(rows, "\n")
}

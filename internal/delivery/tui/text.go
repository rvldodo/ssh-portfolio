package tui

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// para word-wraps text to width w.
func para(s string, w int) string {
	return textStyle.Render(trimLines(lipgloss.Wrap(s, w, "")))
}

// bullet renders "• text" with a hanging indent for wrapped lines.
func bullet(s string, w int) string {
	lines := strings.Split(trimLines(lipgloss.Wrap(s, w-2, "")), "\n")
	for i, l := range lines {
		prefix := "  "
		if i == 0 {
			prefix = accentStyle.Render("•") + " "
		}
		lines[i] = prefix + textStyle.Render(l)
	}
	return strings.Join(lines, "\n")
}

// pills lays out tags as badges, wrapping onto new rows as needed.
func pills(items []string, w int) string {
	var rows []string
	row, rowW := "", 0
	for _, it := range items {
		p := pillStyle.Render(it)
		pw := lipgloss.Width(p)
		if rowW > 0 && rowW+1+pw > w {
			rows = append(rows, row)
			row, rowW = "", 0
		}
		if rowW > 0 {
			row += " "
			rowW++
		}
		row += p
		rowW += pw
	}
	rows = append(rows, row)
	// Blank lines between rows keep the pill backgrounds from touching.
	return strings.Join(rows, "\n\n")
}

// gradient colors each character of s along a blend from a to b.
func gradient(s string, a, b color.Color) string {
	runes := []rune(s)
	colors := lipgloss.Blend1D(len(runes), a, b)
	var out strings.Builder
	for i, r := range runes {
		out.WriteString(lipgloss.NewStyle().Foreground(colors[i]).Bold(true).Render(string(r)))
	}
	return out.String()
}

// truncate shortens s to at most w columns, adding an ellipsis.
func truncate(s string, w int) string {
	if lipgloss.Width(s) <= w {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r))+1 > w {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}

// trimLines removes trailing spaces that wrapping can leave behind; they'd
// push lines one column too wide and misalign the scrollbar.
func trimLines(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	return strings.Join(lines, "\n")
}

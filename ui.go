package main

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// The portfolio is laid out like a website inside a "browser window" card:
//
//	● ● ●   ssh rivaldo.dev/experience          <- address bar (changes per page)
//	 Home   Experience   Projects   Skills  ... <- nav tabs
//	────────────────────────────────────────
//	page content (scrollable viewport)    ┃     <- scrollbar
//	footer key help
//
// Every page renders to a string that goes into one shared viewport, so
// scrolling, the scrollbar, and resizing work the same everywhere.

type page int

const (
	homePage page = iota
	experiencePage
	projectsPage
	skillsPage
	contactPage
	numPages
)

var pageNames = [numPages]string{"Home", "Experience", "Projects", "Skills", "Contact"}
var pagePaths = [numPages]string{"", "/experience", "/projects", "/skills", "/contact"}

const padX = 3 // horizontal padding inside the card border

var (
	colBg      = lipgloss.Color("#0b0b0f")
	colFg      = lipgloss.Color("#a8a8b3")
	colBorder  = lipgloss.Color("236")
	colDim     = lipgloss.Color("243")
	colFaint   = lipgloss.Color("239")
	colBright  = lipgloss.Color("255")
	colAccent  = lipgloss.Color("#2dd4bf")
	colAccent2 = lipgloss.Color("#a78bfa")

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
			Foreground(lipgloss.Color("#0b0b0f")).
			Background(colAccent).
			Bold(true).
			Padding(0, 1)
	ruleStyle = lipgloss.NewStyle().Foreground(colBorder)
)

type model struct {
	page   page
	cursor [numPages]int  // selected row on list pages (Experience, Projects)
	open   [numPages]bool // whether a list page is showing an item's detail
	vp     viewport.Model
	width  int
	height int
}

func newModel() model {
	m := model{vp: viewport.New(), width: 80, height: 24}
	m.vp.KeyMap.Left.SetEnabled(false) // we use ←/→ to switch pages instead
	m.vp.KeyMap.Right.SetEnabled(false)
	m.layout()
	return m
}

func (m model) Init() tea.Cmd { return nil }

// ---------------------------------------------------------------------------
// Update

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
		return m, nil

	case tea.MouseWheelMsg:
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global navigation, like clicking the nav bar.
	switch key {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "tab", "right", "l":
		return m.goTo((m.page + 1) % numPages), nil
	case "shift+tab", "left", "h":
		return m.goTo((m.page + numPages - 1) % numPages), nil
	case "1", "2", "3", "4", "5":
		return m.goTo(page(key[0] - '1')), nil
	}

	// List pages (Experience, Projects) with nothing open: move the selection.
	if n := m.listLen(); n > 0 && !m.open[m.page] {
		switch key {
		case "up", "k":
			m.cursor[m.page] = (m.cursor[m.page] - 1 + n) % n
			m.refresh(false)
			return m, nil
		case "down", "j":
			m.cursor[m.page] = (m.cursor[m.page] + 1) % n
			m.refresh(false)
			return m, nil
		case "enter", "space":
			m.open[m.page] = true
			m.refresh(true)
			return m, nil
		}
	}

	if key == "esc" || key == "backspace" {
		switch {
		case m.open[m.page]: // detail -> back to the list
			m.open[m.page] = false
			m.refresh(true)
		case m.page != homePage: // any page -> home
			return m.goTo(homePage), nil
		default:
			return m, tea.Quit
		}
		return m, nil
	}

	// Everything else (↑/↓, pgup/pgdown, g/G...) scrolls the viewport.
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m model) goTo(p page) model {
	m.page = p
	m.refresh(true)
	return m
}

func (m model) listLen() int {
	switch m.page {
	case experiencePage:
		return len(experiences)
	case projectsPage:
		return len(projects)
	}
	return 0
}

// ---------------------------------------------------------------------------
// Layout

// cardWidth adapts to the terminal: as wide as possible up to a comfortable
// reading width, with a floor so narrow terminals still work.
func (m model) cardWidth() int { return max(48, min(m.width-4, 86)) }

// textWidth is the width available to page content (card minus border,
// padding, and the 2-column scrollbar gutter).
func (m model) textWidth() int { return m.cardWidth() - 2 - padX*2 - 2 }

func (m *model) layout() {
	// Card chrome: border 2 + padding 2 + address bar 1 + gap 1 + tabs 1 +
	// rule 1 + gap 1 + gap 1 + footer 1.
	const chrome = 11
	m.vp.SetWidth(m.textWidth())
	m.vp.SetHeight(max(3, min(m.height-chrome-2, 30)))
	m.refresh(false)
}

// refresh re-renders the current page into the viewport. Call it whenever
// the page, selection, or width changes.
func (m *model) refresh(toTop bool) {
	w := m.textWidth()
	var content string
	selLine := -1 // line of the selected list item, to keep it in view

	switch m.page {
	case homePage:
		content = homeView(w)
	case experiencePage:
		if m.open[m.page] {
			content = experienceDetail(experiences[m.cursor[m.page]], w)
		} else {
			content, selLine = experienceList(m.cursor[m.page], w)
		}
	case projectsPage:
		if m.open[m.page] {
			content = projectDetail(projects[m.cursor[m.page]], w)
		} else {
			content, selLine = projectList(m.cursor[m.page], w)
		}
	case skillsPage:
		content = skillsView(w)
	case contactPage:
		content = contactView(w)
	}

	m.vp.SetContent(content)
	if toTop {
		m.vp.GotoTop()
	}
	if selLine >= 0 {
		// Show the item's two lines; at the top, also show the page heading.
		if selLine <= 3 {
			m.vp.GotoTop()
		}
		m.vp.EnsureVisible(selLine+1, 0, 0)
		m.vp.EnsureVisible(selLine, 0, 0)
	}
}

// ---------------------------------------------------------------------------
// View

func (m model) View() tea.View {
	inner := m.cardWidth() - 2 - padX*2

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(m.textWidth()).MaxWidth(m.textWidth()).Render(m.vp.View()),
		lipgloss.NewStyle().PaddingLeft(1).Render(m.scrollbar()),
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
	v.WindowTitle = profile.Name + " — " + pageNames[m.page]
	v.BackgroundColor = colBg
	v.ForegroundColor = colFg
	return v
}

func (m model) addressBar(width int) string {
	dots := lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render("●") + " " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("●") + " " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("41")).Render("●")
	url := urlStyle.Render(
		faintStyle.Render("ssh ") + brightStyle.Render(domain) + dimStyle.Render(pagePaths[m.page]),
	)
	gap := strings.Repeat(" ", max(1, width-lipgloss.Width(dots)-lipgloss.Width(url)-4)/2)
	return dots + gap + "  " + url
}

func (m model) tabs() string {
	var parts []string
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

func (m model) scrollbar() string {
	vp := m.vp
	h, total := vp.Height(), vp.TotalLineCount()
	track := ruleStyle.Render("│")
	if total <= h {
		return strings.TrimSuffix(strings.Repeat(" \n", h), "\n") // no bar needed
	}
	thumb := lipgloss.NewStyle().Foreground(colDim).Render("┃")
	thumbH := max(1, h*h/total)
	thumbTop := int(float64(h-thumbH) * vp.ScrollPercent())
	rows := make([]string, h)
	for i := range rows {
		rows[i] = track
		if i >= thumbTop && i < thumbTop+thumbH {
			rows[i] = thumb
		}
	}
	return strings.Join(rows, "\n")
}

// ---------------------------------------------------------------------------
// Pages. Each returns plain styled text that fits within width w.

func homeView(w int) string {
	var b strings.Builder
	b.WriteString(gradient(profile.Name, colAccent, colAccent2) + "\n")
	b.WriteString(dimStyle.Render(profile.Title+" · "+profile.Location) + "\n\n")
	b.WriteString(para(profile.Summary, w) + "\n\n")

	b.WriteString(h2Style.Render("At a glance") + "\n\n")
	for _, g := range profile.Glance {
		b.WriteString(brightStyle.Width(18).Render(g[0]) + textStyle.Render(g[1]) + "\n")
	}

	b.WriteString("\n" + h2Style.Render("Now") + "\n\n")
	b.WriteString(para(profile.Now, w) + "\n\n")

	b.WriteString(h2Style.Render("Education") + "\n\n")
	for _, e := range education {
		b.WriteString(brightStyle.Render(e.School) + "\n")
		b.WriteString(dimStyle.Render(e.Field+" · "+e.Place+" · "+e.Year) + "\n")
	}

	b.WriteString("\n" + faintStyle.Render("→ press 2 for experience, 3 for projects"))
	return b.String()
}

// experienceList returns the rendered list and the line of the selected item.
func experienceList(cursor, w int) (string, int) {
	var rows [][2]string
	for _, e := range experiences {
		rows = append(rows, [2]string{e.Role + " · " + e.Company, e.Period + " · " + e.Type})
	}
	return list("Experience", rows, cursor, w)
}

func projectList(cursor, w int) (string, int) {
	var rows [][2]string
	for _, p := range projects {
		rows = append(rows, [2]string{p.Name, p.Kind + " · " + p.Period})
	}
	return list("Selected projects", rows, cursor, w)
}

func list(title string, rows [][2]string, cursor, w int) (string, int) {
	lines := []string{h1Style.Render(title), ""}
	sel := 0
	for i, r := range rows {
		marker, style := "  ", textStyle
		if i == cursor {
			marker, style, sel = accentStyle.Render("› "), selStyle, len(lines)
		}
		lines = append(lines,
			marker+style.Render(truncate(r[0], w-2)),
			"  "+dimStyle.Render(truncate(r[1], w-2)),
			"")
	}
	return strings.Join(lines, "\n"), sel
}

func experienceDetail(e experience, w int) string {
	var b strings.Builder
	b.WriteString(h1Style.Render(e.Role) + "\n")
	b.WriteString(accentStyle.Render(e.Company) + dimStyle.Render(" · "+e.Org) + "\n")
	b.WriteString(dimStyle.Render(e.Period+" · "+e.Type+" · "+e.Place) + "\n\n")
	if e.About != "" {
		b.WriteString(para(e.About, w) + "\n\n")
	}
	b.WriteString(h2Style.Render("What I did") + "\n\n")
	for _, s := range e.Bullets {
		b.WriteString(bullet(s, w) + "\n")
	}
	b.WriteString("\n" + h2Style.Render("Stack") + "\n\n" + pills(e.Stack, w))
	return b.String()
}

func projectDetail(p project, w int) string {
	var b strings.Builder
	b.WriteString(h1Style.Render(p.Name) + "\n")
	b.WriteString(accentStyle.Render(p.Kind) + "\n")
	b.WriteString(dimStyle.Render(p.Period+" · "+p.Place) + "\n\n")
	b.WriteString(para(p.About, w) + "\n\n")
	b.WriteString(h2Style.Render("Highlights") + "\n\n")
	for _, s := range p.Bullets {
		b.WriteString(bullet(s, w) + "\n")
	}
	b.WriteString("\n" + h2Style.Render("Stack") + "\n\n" + pills(p.Stack, w))
	return b.String()
}

func skillsView(w int) string {
	var b strings.Builder
	b.WriteString(h1Style.Render("Skills") + "\n")
	for _, s := range skills {
		b.WriteString("\n" + h2Style.Render(s.Group) + "\n" + pills(s.Items, w) + "\n")
	}
	return b.String()
}

func contactView(w int) string {
	var b strings.Builder
	b.WriteString(h1Style.Render("Let's talk") + "\n\n")
	b.WriteString(para("I'm always happy to chat about backend systems, payment integrations, "+
		"or shipping reliable software. The fastest way to reach me is email.", w) + "\n\n")
	for _, c := range contacts {
		b.WriteString(dimStyle.Width(10).Render(c[0]) + accentStyle.Render(c[1]) + "\n")
	}
	b.WriteString("\n" + faintStyle.Render("Thanks for visiting. Press q to leave."))
	return b.String()
}

// ---------------------------------------------------------------------------
// Small rendering helpers

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

// pills lays out tags as little badges, wrapping to new lines as needed.
func pills(items []string, w int) string {
	var lines []string
	line, lineW := "", 0
	for _, it := range items {
		p := pillStyle.Render(it)
		pw := lipgloss.Width(p)
		if lineW > 0 && lineW+1+pw > w {
			lines = append(lines, line)
			line, lineW = "", 0
		}
		if lineW > 0 {
			line += " "
			lineW++
		}
		line += p
		lineW += pw
	}
	lines = append(lines, line)
	// Blank lines between rows of pills keep the backgrounds from touching.
	return strings.Join(lines, "\n\n")
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

// trimLines removes trailing spaces that wrapping can leave behind.
func trimLines(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	return strings.Join(lines, "\n")
}

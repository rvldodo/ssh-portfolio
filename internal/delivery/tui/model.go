package tui

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"ssh-portfolio/internal/domain"
)

type page int

const (
	homePage page = iota
	experiencePage
	projectsPage
	skillsPage
	contactPage
	numPages
)

var (
	pageNames = [numPages]string{"Home", "Experience", "Projects", "Skills", "Contact"}
	pagePaths = [numPages]string{"", "/experience", "/projects", "/skills", "/contact"}
)

// model is the Bubble Tea state for one SSH session.
//
// Every page renders to a string that goes into one shared viewport, so
// scrolling, the scrollbar, and resizing work the same everywhere.
type model struct {
	data Portfolio
	exps []domain.Experience // fetched once per session; list pages index into these
	prjs []domain.Project

	page   page
	cursor [numPages]int  // selected row on list pages (Experience, Projects)
	open   [numPages]bool // whether a list page is showing an item's detail
	vp     viewport.Model
	width  int
	height int
}

// New returns the TUI for one session, reading content from data.
func New(data Portfolio) tea.Model {
	m := model{
		data:   data,
		exps:   data.Experiences(),
		prjs:   data.Projects(),
		vp:     viewport.New(),
		width:  80,
		height: 24,
	}
	m.vp.KeyMap.Left.SetEnabled(false) // ←/→ switch pages instead
	m.vp.KeyMap.Right.SetEnabled(false)
	m.layout()
	return m
}

func (m model) Init() tea.Cmd { return nil }

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

	// List pages with nothing open: move the selection or open an item.
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
			return m, nil
		case m.page != homePage: // any page -> home
			return m.goTo(homePage), nil
		default:
			return m, tea.Quit
		}
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
		return len(m.exps)
	case projectsPage:
		return len(m.prjs)
	}
	return 0
}

// cardWidth adapts to the terminal: as wide as possible up to a comfortable
// reading width, with a floor so narrow terminals still work.
func (m model) cardWidth() int { return max(48, min(m.width-4, 86)) }

// textWidth is the width for page content: the card minus border, padding,
// and the 2-column scrollbar gutter.
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
		content = homeView(m.data, w)
	case experiencePage:
		if m.open[m.page] {
			content = experienceDetail(m.exps[m.cursor[m.page]], w)
		} else {
			content, selLine = experienceList(m.exps, m.cursor[m.page], w)
		}
	case projectsPage:
		if m.open[m.page] {
			content = projectDetail(m.prjs[m.cursor[m.page]], w)
		} else {
			content, selLine = projectList(m.prjs, m.cursor[m.page], w)
		}
	case skillsPage:
		content = skillsView(m.data.Skills(), w)
	case contactPage:
		content = contactView(m.data.Contacts(), w)
	}

	m.vp.SetContent(content)
	if toTop {
		m.vp.GotoTop()
	}
	if selLine >= 0 {
		// Show both lines of the item; near the top, also show the heading.
		if selLine <= 3 {
			m.vp.GotoTop()
		}
		m.vp.EnsureVisible(selLine+1, 0, 0)
		m.vp.EnsureVisible(selLine, 0, 0)
	}
}

package tui

import (
	"fmt"
	"strings"

	"ssh-portfolio/internal/domain"
)

// Each function renders one page as styled text that fits within width w.
// They're pure functions of their inputs, so they're easy to test and tweak.

func homeView(data Portfolio, w int) string {
	p := data.Profile()
	var b strings.Builder

	sub := p.Title + " · " + p.Location
	if years := data.YearsOfExperience(); years > 0 {
		sub += fmt.Sprintf(" · %d+ years", years)
	}
	b.WriteString(gradient(p.Name, colAccent, colAccent2) + "\n")
	b.WriteString(dimStyle.Render(sub) + "\n\n")
	b.WriteString(para(p.Summary, w) + "\n\n")

	if len(p.Highlights) > 0 {
		b.WriteString(h2Style.Render("At a glance") + "\n\n")
		for _, h := range p.Highlights {
			b.WriteString(brightStyle.Width(18).Render(h.Label) + textStyle.Render(h.Detail) + "\n")
		}
		b.WriteString("\n")
	}

	if p.Now != "" {
		b.WriteString(h2Style.Render("Now") + "\n\n" + para(p.Now, w) + "\n\n")
	}

	if edu := data.Education(); len(edu) > 0 {
		b.WriteString(h2Style.Render("Education") + "\n\n")
		for _, e := range edu {
			b.WriteString(brightStyle.Render(e.School) + "\n")
			b.WriteString(
				dimStyle.Render(fmt.Sprintf("%s · %s · %d", e.Field, e.Location, e.Year)) + "\n",
			)
		}
		b.WriteString("\n")
	}

	b.WriteString(faintStyle.Render("→ press 2 for experience, 3 for projects"))
	return b.String()
}

// experienceList returns the rendered list and the line of the selected item.
func experienceList(exps []domain.Experience, cursor, w int) (string, int) {
	rows := make([][2]string, len(exps))
	for i, e := range exps {
		rows[i] = [2]string{e.Role + " · " + e.Company, e.Period.String() + " · " + e.Employment}
	}
	return list("Experience", rows, cursor, w)
}

func projectList(prjs []domain.Project, cursor, w int) (string, int) {
	rows := make([][2]string, len(prjs))
	for i, p := range prjs {
		rows[i] = [2]string{p.Name, p.Kind + " · " + p.Period.String()}
	}
	return list("Selected projects", rows, cursor, w)
}

// list renders two-line rows with a › marker on the selected one.
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

func experienceDetail(e domain.Experience, w int) string {
	var b strings.Builder
	b.WriteString(h1Style.Render(e.Role) + "\n")
	b.WriteString(accentStyle.Render(e.Company) + dimStyle.Render(" · "+e.Org) + "\n")
	b.WriteString(dimStyle.Render(e.Period.String()+" · "+e.Employment+" · "+e.Location) + "\n\n")
	if e.About != "" {
		b.WriteString(para(e.About, w) + "\n\n")
	}
	b.WriteString(h2Style.Render("What I did") + "\n\n")
	for _, s := range e.Achievements {
		b.WriteString(bullet(s, w) + "\n")
	}
	b.WriteString("\n" + h2Style.Render("Stack") + "\n\n" + pills(e.Stack, w))
	return b.String()
}

func projectDetail(p domain.Project, w int) string {
	var b strings.Builder
	b.WriteString(h1Style.Render(p.Name) + "\n")
	b.WriteString(accentStyle.Render(p.Kind) + "\n")
	b.WriteString(dimStyle.Render(p.Period.String()+" · "+p.Location) + "\n\n")
	if p.About != "" {
		b.WriteString(para(p.About, w) + "\n\n")
	}
	b.WriteString(h2Style.Render("Highlights") + "\n\n")
	for _, s := range p.Highlights {
		b.WriteString(bullet(s, w) + "\n")
	}
	b.WriteString("\n" + h2Style.Render("Stack") + "\n\n" + pills(p.Stack, w))
	return b.String()
}

func skillsView(groups []domain.SkillGroup, w int) string {
	var b strings.Builder
	b.WriteString(h1Style.Render("Skills") + "\n")
	for _, g := range groups {
		b.WriteString("\n" + h2Style.Render(g.Name) + "\n" + pills(g.Skills, w) + "\n")
	}
	return b.String()
}

func contactView(contacts []domain.Contact, w int) string {
	var b strings.Builder
	b.WriteString(h1Style.Render("Let's talk") + "\n\n")
	b.WriteString(para("I'm always happy to chat about backend systems, payment integrations, "+
		"or shipping reliable software. The fastest way to reach me is email.", w) + "\n\n")
	for _, c := range contacts {
		b.WriteString(dimStyle.Width(10).Render(c.Label) + accentStyle.Render(c.Value) + "\n")
	}
	b.WriteString("\n" + faintStyle.Render("Thanks for visiting. Press q to leave."))
	return b.String()
}

package domain

import "time"

// Period is a span of months. A nil End means the period is ongoing.
type Period struct {
	Start time.Time
	End   *time.Time
}

// Ongoing reports whether the period has no end yet.
func (p Period) Ongoing() bool { return p.End == nil }

// String formats the period as "Sep 2024 – Present" or "Apr 2023 – Aug 2023".
func (p Period) String() string {
	end := "Present"
	if p.End != nil {
		end = p.End.Format("Jan 2006")
	}
	return p.Start.Format("Jan 2006") + " – " + end
}

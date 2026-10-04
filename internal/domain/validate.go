package domain

import (
	"errors"
	"fmt"
)

// Validate checks the business rules a portfolio must satisfy before it can
// be shown. It reports every problem at once rather than stopping at the
// first, so content mistakes can be fixed in one pass.
func (p Portfolio) Validate() error {
	var errs []error
	if p.Profile.Name == "" {
		errs = append(errs, errors.New("profile: name is required"))
	}
	if p.Site.Address == "" {
		errs = append(errs, errors.New("site: address is required"))
	}

	seen := map[string]bool{}
	for i, e := range p.Experiences {
		where := fmt.Sprintf("experiences[%d] (%s)", i, e.Slug)
		errs = append(errs, checkSlug(where, e.Slug, seen)...)
		if e.Company == "" || e.Role == "" {
			errs = append(errs, fmt.Errorf("%s: company and role are required", where))
		}
		errs = append(errs, checkPeriod(where, e.Period)...)
	}

	seen = map[string]bool{}
	for i, pr := range p.Projects {
		where := fmt.Sprintf("projects[%d] (%s)", i, pr.Slug)
		errs = append(errs, checkSlug(where, pr.Slug, seen)...)
		if pr.Name == "" {
			errs = append(errs, fmt.Errorf("%s: name is required", where))
		}
		errs = append(errs, checkPeriod(where, pr.Period)...)
	}

	return errors.Join(errs...)
}

func checkSlug(where, slug string, seen map[string]bool) []error {
	switch {
	case slug == "":
		return []error{fmt.Errorf("%s: slug is required", where)}
	case seen[slug]:
		return []error{fmt.Errorf("%s: duplicate slug %q", where, slug)}
	}
	seen[slug] = true
	return nil
}

func checkPeriod(where string, p Period) []error {
	if p.Start.IsZero() {
		return []error{fmt.Errorf("%s: period start is required", where)}
	}
	if p.End != nil && p.End.Before(p.Start) {
		return []error{fmt.Errorf("%s: period ends before it starts", where)}
	}
	return nil
}

package wiki

import "github.com/cooper/quiki/wikifier"

// generate.go - deferred validation and pregeneration support

// Check represents a validation check to retry later
type Check struct {
	PageName string
	Func     func() bool // returns true if validation now passes
	Warning  string
	Pos      wikifier.Position
}

// AddCheck adds a validation check - either executes immediately or defers based on pregeneration mode
func (w *Wiki) AddCheck(pageName, warning string, pos wikifier.Position, validateFunc func() bool) {
	// if validation passes immediately, no need to defer or warn
	if validateFunc() {
		return
	}

	// validation failed - either defer it or warn immediately
	if w.pregenerating {
		// defer the check for later
		w.checkMu.Lock()
		defer w.checkMu.Unlock()

		w.checks = append(w.checks, Check{
			PageName: pageName,
			Func:     validateFunc,
			Warning:  warning,
			Pos:      pos,
		})
	} else {
		// not pregenerating, add warning immediately
		page := w.FindPage(pageName)
		if page.Exists() {
			pageWarn(page, warning, pos)
		}
	}
}

// ProcessChecks runs all deferred validation checks and adds warnings for those that still fail
func (w *Wiki) ProcessChecks() {
	w.checkMu.Lock()
	defer w.checkMu.Unlock()

	for _, dv := range w.checks {
		if !dv.Func() {
			// still fails, add the warning to the page
			page := w.FindPage(dv.PageName)
			if page.Exists() {
				warning := wikifier.Warning{Message: dv.Warning, Pos: dv.Pos}
				page.Warnings = append(page.Warnings, warning)
			}
		}
	}

	w.checks = nil
}

// SetDeferringChecks sets whether we're postponing checks until after a bulk operation (e.g., pregenerating)
func (w *Wiki) SetDeferringChecks(scoped bool) {
	w.pregenerating = scoped
}

// IsDeferringChecks returns whether we're currently postponing checks
func (w *Wiki) IsDeferringChecks() bool {
	return w.pregenerating
}

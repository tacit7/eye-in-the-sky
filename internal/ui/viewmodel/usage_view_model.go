package viewmodel

import "time"

// Section represents a rendered section with title, rows, and footer
type Section struct {
	Title   string
	Rows    []string // built by a mapper that knows current width
	Footer  map[string]string
	HasData bool
}

// UsageViewModel holds precomputed view state for the usage tab
type UsageViewModel struct {
	Sections    []Section
	TotalsUSD   float64
	TokensTotal int
	LastSync    time.Time
	HasData     bool
}

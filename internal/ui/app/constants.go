package app

import "time"

// Layout constants
const (
	MinHeaderWidth     = 40
	HeaderPadding      = 2
	FooterPadding      = 2
	ContentBoxPadding  = 4
	StatusTimeout      = 3 * time.Second
	MinVisibleHeight   = 10
	HeaderReservedRows = 3
	TabsReservedRows   = 2
	FooterReservedRows = 2
)

// Time display constants
const (
	JustNowThreshold = 1 * time.Second
	SecondsThreshold = 60 * time.Second
	MinutesThreshold = 60 * time.Minute
	HoursThreshold   = 24 * time.Hour
)

// UI component constants
const (
	AppVersion         = "v0.2.0"
	AppTitle           = "\uf06e Eye in the Sky - Agent Management"
	StatusMessageWidth = 40
)

// Layout widths
const (
	SplitPaneLeftRatio = 3 // Left pane takes 1/3 of width
)

// Pagination constants
const (
	PageSize = 10
)
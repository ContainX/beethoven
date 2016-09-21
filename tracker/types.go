package tracker

import (
	"time"
)

type Updates struct {
	LastSync           time.Time
	LastConfigRendered time.Time
	LastConfigValid    time.Time
	LastProxyReload    time.Time
}

type Status struct {
	LastUpdated     Updates
	LastError       error
	ValidationError *ValidationError
}

type ValidationError struct {
	Error        error
	FailedConfig string
}

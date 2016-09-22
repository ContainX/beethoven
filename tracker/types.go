package tracker

import (
	"time"
)

type Updates struct {
	LastSync           time.Time `json:"last_sync"`
	LastConfigRendered time.Time `json:"last_config_rendered"`
	LastConfigValid    time.Time `json:"last_config_valid"`
	LastProxyReload    time.Time `json:"last_proxy_reload"`
}

type Status struct {
	LastUpdated     Updates          `json:"last_updated"`
	LastError       error            `json:"last_error"`
	ValidationError *ValidationError `json:"validation_error"`
}

type ValidationError struct {
	Error        error  `json:"error"`
	FailedConfig string `json:"failed_config"`
}

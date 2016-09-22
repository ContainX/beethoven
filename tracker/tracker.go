package tracker

import (
	"github.com/ContainX/beethoven/config"
	"time"
)

// Tracker is responsible for keeping track of state and updates
// throughout Beethoven.  It serves as a common information hub to
// the API
type Tracker struct {
	cfg    *config.Config
	status Status
}

func New(cfg *config.Config) *Tracker {
	return &Tracker{
		cfg: cfg,
		status: Status{
			LastUpdated: Updates{},
		},
	}
}

func (tr *Tracker) GetStatus() Status {
	return tr.status
}

func (tr *Tracker) SetError(err error) {
	tr.status.LastError = err
}

func (tr *Tracker) SetValidationError(verr *ValidationError) {
	tr.status.ValidationError = verr
}

func (tr *Tracker) ClearValidationError() {
	tr.status.ValidationError = nil
}

// SetLastSync will set the time we fetched a snapshot from Marathon
func (tr *Tracker) SetLastSync(t time.Time) {
	tr.status.LastUpdated.LastSync = t
}

// SetLastConfigRendered will set the time we rendered a temporary config
func (tr *Tracker) SetLastConfigRendered(t time.Time) {
	tr.status.LastUpdated.LastConfigRendered = t
}

// SetLastConfigValid captures the last time we had a successful rendered config validate
// via the proxy
func (tr *Tracker) SetLastConfigValid(t time.Time) {
	tr.status.LastUpdated.LastConfigValid = t
}

// SetLastProxyReload the last time we executed a reload on the proxy
func (tr *Tracker) SetLastProxyReload(t time.Time) {
	tr.status.LastUpdated.LastProxyReload = t
}

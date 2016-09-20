package tracker

import (
	"github.com/ContainX/beethoven/config"
	"time"
)

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

func (tr *Tracker) SetError(err error) {
	tr.status.LastError = err
}

func (tr *Tracker) SetLastSync(t time.Time) {
	tr.status.LastUpdated.LastSync = t
}

func (tr *Tracker) SetLastConfigRendered(t time.Time) {
	tr.status.LastUpdated.LastConfigRendered = t
}

func (tr *Tracker) SetLastConfigValid(t time.Time) {
	tr.status.LastUpdated.LastConfigValid = t
}

func (tr *Tracker) SetLastProxyReload(t time.Time) {
	tr.status.LastUpdated.LastProxyReload = t
}

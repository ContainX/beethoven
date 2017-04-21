package scheduler

import (
	"github.com/ContainX/beethoven/tracker"
	"github.com/ContainX/beethoven/config"
	"github.com/ContainX/depcon/pkg/logger"
	"strings"
)

type schedulerService struct {
	cfg          *config.Config
	tracker      *tracker.Tracker
	handler      func(proxyConf string)
}

var (
	log = logger.GetLogger("beethoven.scheduler")
)

func NewScheduler(cfg *config.Config, tracker *tracker.Tracker) Scheduler {
	if cfg.SchedulerType == config.MarathonScheduler {
		return &marathonService{cfg, tracker}
	}

	return &swarmService{cfg, tracker}
}

func (s *schedulerService) shouldTriggerReload(appId string, event interface{}) bool {
	if appId == "" {
		log.Warning("Event: Could not locate AppId: %s", event)
		return false
	}

	trigger := true

	if s.cfg.IsFilterDefined() {
		match := s.cfg.Filter().MatchString(appId)
		log.Debug("Matching appId: %s to filter: %s -> %v", appId, s.cfg.FilterRegExStr, match)
		log.Debug("Event: %s", event)
		trigger = match
	}

	return trigger
}

// Change IDs using /'s to '-' since we need identifiers
// that are compat with templates.
// ex: /products/stores/someservice would be products-stores-someservice
func appIdToDashes(appId string) string {
	parts := strings.Split(appId[1:], "/")
	return strings.Join(parts, "-")
}



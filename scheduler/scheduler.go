package scheduler

import (
	"github.com/ContainX/beethoven/config"
	"github.com/ContainX/beethoven/tracker"
	"github.com/ContainX/depcon/pkg/logger"
)

type schedulerService struct {
	cfg     *config.Config
	tracker *tracker.Tracker
	reload  chan bool
}

var (
	log = logger.GetLogger("beethoven.scheduler")
)

func NewScheduler(cfg *config.Config, tracker *tracker.Tracker) Scheduler {

	ss := &schedulerService{cfg: cfg, tracker: tracker}

	switch cfg.SchedulerType {
	case config.MarathonScheduler:
		return createMarathonScheduler(ss)
	default:
		return createSwarmScheduler(ss)
	}
}

func (s *schedulerService) shouldTriggerReload(appId string, event interface{}) bool {
	if appId == "" {
		log.Warningf("Event: Could not locate AppId: %s", event)
		return false
	}

	trigger := true

	if s.cfg.IsFilterDefined() {
		trigger = s.cfg.Filter().MatchString(appId)
		log.Debugf("Matching appId: %s to filter: %s -> %v, Event: %s", appId, s.cfg.FilterRegExStr, trigger, event)
	}
	return trigger
}

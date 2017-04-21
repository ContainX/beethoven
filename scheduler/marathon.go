package scheduler

import (
	"github.com/ContainX/depcon/marathon"
	"github.com/ContainX/depcon/pkg/logger"
	"time"
)

type marathonService struct {
	*schedulerService
	events       marathon.EventsChannel
	marathon     marathon.Marathon
	shutdown     ShutdownChan
	reloadQueue  ReloadChan
}

func (m *marathonService) init() {
	m.shutdown = make(ShutdownChan, 2)
	m.reloadQueue = make(ReloadChan, 2)
}

// Watch for changes using streams and make callbacks to the specified
// handler when apps have been added, removed or health changes.
func (m *marathonService) Watch(handler func(proxyConf string)) {
	m.handler = handler

	// MVP - no health checks - should verify and use healthy masters
	m.marathon = marathon.NewMarathonClient(m.cfg.MarthonUrls[0], m.cfg.Username, m.cfg.Password)

	// suppress marathon debug
	logger.SetLevel(logger.WARNING, "client")
	logger.SetLevel(logger.WARNING, "depcon.marathon")
}

// Shutdown the current stream watching
func (m *marathonService) Shutdown() {
	m.shutdown <- true
}

// Fetch all applications/services from the scheduler source
func (m *marathonService) FetchApps() (map[string]*App, error) {
	apps, err := m.marathon.ListApplicationsWithFilters("embed=apps.tasks")
	if err != nil {
		log.Error("Error fetching apps: %s", err.Error())
		return nil, err
	}

	result := map[string]*App{}

	for _, a := range apps.Apps {

		// Create template based app
		tapp := new(App)
		tapp.AppId = appIdToDashes(a.ID)
		tapp.Env = a.Env
		tapp.Labels = a.Labels
		tapp.Tasks = []Task{}

		// Iterate through the apps tasks - remove any tasks that do not match
		// our criteria for being healthy
		for _, t := range a.Tasks {
			// Skip tasks with no ports
			if len(t.Ports) == 0 {
				continue
			}

			// If health checks defined, make sure they are passing
			if a.HealthChecks != nil && len(a.HealthChecks) > 0 {
				if t.HealthCheckResult == nil || len(t.HealthCheckResult) == 0 {
					// currently deploying - no health checks yet
					continue
				}

				// If any health check is failing lets assume failure altogether
				healthy := true
				for _, hc := range t.HealthCheckResult {
					if hc.Alive == false {
						healthy = false
					}
				}

				if healthy == false {
					continue
				}
			}
			tapp.Tasks = append(tapp.Tasks, marathonTaskToTask(t))
		}

		// Only add apps with tasks
		if len(tapp.Tasks) > 0 {
			result[tapp.AppId] = tapp
		}

	}
	m.tracker.SetLastSync(time.Now())
	return result, nil
}

func (m *marathonService) initSSEStream() {
	m.events = make(marathon.EventsChannel, 5)

	filter := marathon.EventIDStatusUpdate | marathon.EventIDChangedHealthCheck

	err := m.marathon.CreateEventStreamListener(m.events, filter)
	if err != nil {
		log.Fatalf("Failed to register for events, %s", err)
	}

	go m.streamListener()
}

func (m *marathonService) streamListener() {
	stop := false
	for {
		if stop {
			break
		}
		select {
		case <-m.shutdown:
			stop = true
		case event := <-m.events:
			{
				// If the event doesn't contain the AppId -- skip
				appId := m.getAppID(event)
				if m.shouldTriggerReload(appId, event) {
					select {
					case m.reloadQueue <- true:
					default:
						log.Warning("Reload queue is full")
					}
				}
			}
		}
	}
	m.marathon.CloseEventStreamListener(m.events)
}

// getAppID returns the application indentifier for only the evens we care to
// trigger a configuration reload from.
func (m *marathonService) getAppID(e *marathon.Event) string {
	switch e.ID {
	case marathon.EventIDStatusUpdate:
		return toEventStatusUpdate(e).AppID
	case marathon.EventIDAPIRequest:
		return toEventAPIRequest(e).AppDefinition.ID
	case marathon.EventIDChangedHealthCheck:
		return toEventHealthCheckChanged(e).AppID
	}
	return ""
}

func toEventStatusUpdate(e *marathon.Event) *marathon.EventStatusUpdate {
	return e.Event.(*marathon.EventStatusUpdate)
}

func toEventAPIRequest(e *marathon.Event) *marathon.EventAPIRequest {
	return e.Event.(*marathon.EventAPIRequest)
}

func toEventHealthCheckChanged(e *marathon.Event) *marathon.EventHealthCheckChanged {
	return e.Event.(*marathon.EventHealthCheckChanged)
}

func marathonTaskToTask(mt *marathon.Task) Task {
	task := Task{}
	task.Host = mt.Host
	task.Ports = mt.Ports
	task.ServicePorts = mt.ServicePorts
	task.StagedAt = mt.StagedAt
	task.StartedAt = mt.StagedAt
	task.Version = mt.Version
	return task
}

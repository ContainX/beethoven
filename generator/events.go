package generator

import (
	"github.com/ContainX/depcon/marathon"
)

func (g *Generator) initSSEStream() {
	g.events = make(marathon.EventsChannel, 5)

	filter := marathon.EventIDStatusUpdate | marathon.EventIDChangedHealthCheck

	err := g.marathon.CreateEventStreamListener(g.events, filter)
	if err != nil {
		log.Fatalf("Failed to register for events, %s", err)
	}

	go g.streamListener()
}

func (g *Generator) streamListener() {
	stop := false
	for {
		if stop {
			break
		}
		select {
		case <-g.shutdown:
			stop = true
		case event := <-g.events:
			{
				// If the event doesn't contain the AppId -- skip
				appId := g.getAppID(event)
				if g.shouldTriggerReload(appId, event) {
					select {
					case g.reloadQueue <- true:
					default:
						log.Warning("Reload queue is full")
					}
				}
			}
		}
	}
	g.marathon.CloseEventStreamListener(g.events)
}

func (g *Generator) shouldTriggerReload(appId string, event *marathon.Event) bool {
	if appId == "" {
		log.Warning("Event: Could not locate AppId: %s", event)
		return false
	}

	trigger := true

	if g.cfg.IsFilterDefined() {
		match := g.cfg.Filter().MatchString(appId)
		log.Debug("Matching appId: %s to filter: %s -> %v", appId, g.cfg.FilterRegExStr, match)
		log.Debug("Event: %s", event)
		trigger = match
	}

	return trigger
}

// getAppID returns the application indentifier for only the evens we care to
// trigger a configuration reload from.
func (g *Generator) getAppID(e *marathon.Event) string {
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

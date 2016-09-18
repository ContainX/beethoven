package generator

import (
	"github.com/ContainX/beethoven/config"
	"github.com/ContainX/depcon/marathon"
	"github.com/ContainX/depcon/pkg/logger"
)

type Generator struct {
	cfg      *config.Config
	handler  func(proxyConf string)
	events   marathon.EventsChannel
	marathon marathon.Marathon
	shutdown ShutdownChan
}

type ShutdownChan chan bool

var (
	log = logger.GetLogger("beethoven")
)

func New(cfg *config.Config) *Generator {
	return &Generator{
		cfg:      cfg,
		shutdown: make(ShutdownChan, 2),
	}
}

func (g *Generator) Watch(handler func(proxyConf string)) {
	g.handler = handler

	// MVP - no health checks - should verify and use healthy masters
	g.marathon = marathon.NewMarathonClient(g.cfg.MarthonUrls[0], g.cfg.Username, g.cfg.Password)
	g.events = make(marathon.EventsChannel, 5)

	filter := marathon.EventIDStatusUpdate | marathon.EventIDAPIRequest | marathon.EventIDChangedHealthCheck

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
			log.Info("Received event: %s", event)
		}
	}
	g.marathon.CloseEventStreamListener(g.events)
}

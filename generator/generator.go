package generator

import (
	"github.com/ContainX/beethoven/config"
	"github.com/ContainX/beethoven/scheduler"
	"github.com/ContainX/beethoven/tracker"
	"github.com/ContainX/depcon/pkg/logger"
	"time"
)

type Generator struct {
	cfg          *config.Config
	tracker      *tracker.Tracker
	scheduler    scheduler.Scheduler
	reloadQueue  ReloadChan
	handler      func(proxyConf string)
	templateData TemplateData
}

type ReloadChan chan bool

var (
	log = logger.GetLogger("beethoven.generator")
)

func New(cfg *config.Config, tracker *tracker.Tracker, scheduler scheduler.Scheduler) *Generator {
	return &Generator{
		cfg:          cfg,
		tracker:      tracker,
		reloadQueue:  make(chan bool, 2),
		scheduler:    scheduler,
		templateData: TemplateData{},
	}
}

// Watch marathon for changes using streams and make callbacks to the specified
// handler when apps have been added, removed or health changes.
func (g *Generator) Watch(handler func(proxyConf string)) {
	g.handler = handler

	g.scheduler.Watch(g.reloadQueue)
	g.generateConfig()
	go g.initReloadWatcher()
}

// Watches the reload channel and generated a new config
func (g *Generator) initReloadWatcher() {
	throttle := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-throttle.C:
			<-g.reloadQueue
			log.Info("configuration reload triggered")
			g.generateConfig()
		}
	}
}

func (g *Generator) ReloadConfiguration() {
	g.generateConfig()
}

func (g *Generator) generateConfig() {
	if apps, err := g.scheduler.FetchApps(); err != nil {
		log.Error("Skipping config generation...")
		g.tracker.SetError(err)
		return
	} else {
		g.templateData.Apps = apps
	}

	changed, err := g.writeConfiguration()
	if err != nil {
		log.Error(err.Error())
		g.tracker.SetError(err)
		return
	}

	if changed {
		log.Info("Reloading NGINX")
		err = g.reload()
		if err != nil {
			log.Error(err.Error())
			g.tracker.SetError(err)
			return
		}
	}

	// No errors - clear tracker
	g.tracker.SetError(nil)
}

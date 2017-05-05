package proxy

import (
	"fmt"
	"github.com/ContainX/beethoven/config"
	"github.com/ContainX/beethoven/generator"
	"github.com/ContainX/beethoven/scheduler"
	"github.com/ContainX/beethoven/tracker"
	"github.com/ContainX/depcon/pkg/logger"
	"github.com/gorilla/mux"
	"net/http"
)

const (
	versionResponse = `{
  "name": "beethoven",
  "description": "Mesos/Marathon HTTP Proxy",
  "version": "%s"
}`
)

var (
	log = logger.GetLogger("beethoven.proxy")
)

type Proxy struct {
	cfg        *config.Config
	httpServer *http.Server
	generator  *generator.Generator
	scheduler  scheduler.Scheduler
	tracker    *tracker.Tracker
	mux        *mux.Router
}

func New(cfg *config.Config) *Proxy {
	//cfg.SchedulerType = config.SwarmScheduler
	return &Proxy{
		cfg: cfg,
	}
}

func (p *Proxy) initRoutes() {
	p.mux.HandleFunc("/bt", p.getVersion)
	p.mux.HandleFunc("/bt/status/", p.getStatus)
	p.mux.HandleFunc("/bt/config/", p.getConfig)
	p.mux.HandleFunc("/bt/reload/", p.reloadConfig)
	p.mux.HandleFunc("/bt/reloadall/", p.reloadAll)

}

func (p *Proxy) Serve() {
	p.mux = mux.NewRouter()
	p.initRoutes()

	p.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", p.cfg.HttpPort()),
		Handler: p.mux,
	}

	p.tracker = tracker.New(p.cfg)
	p.scheduler = scheduler.NewScheduler(p.cfg, p.tracker)

	// Start  configuration generator
	p.generator = generator.New(p.cfg, p.tracker, p.scheduler)
	p.generator.Watch(p.debugConfig)

	p.httpServer.ListenAndServe()
}

func (p *Proxy) debugConfig(conf string) {
}

func (p *Proxy) getVersion(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, versionResponse, p.cfg.Version)
}

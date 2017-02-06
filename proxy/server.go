package proxy

import (
	"fmt"
	"github.com/ContainX/beethoven/config"
	"github.com/ContainX/beethoven/generator"
	"github.com/ContainX/beethoven/tracker"
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

type Proxy struct {
	cfg        *config.Config
	httpServer *http.Server
	generator  *generator.Generator
	tracker    *tracker.Tracker
	mux        *mux.Router
}

func New(cfg *config.Config) *Proxy {
	return &Proxy{
		cfg: cfg,
	}
}

func (p *Proxy) initRoutes() {
	p.mux.HandleFunc("/bt", p.getVersion)
	p.mux.HandleFunc("/bt/status/", p.getStatus)
	p.mux.HandleFunc("/bt/config/", p.getConfig)
}

func (p *Proxy) Serve() {
	p.mux = mux.NewRouter()
	p.initRoutes()

	p.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", p.cfg.HttpPort()),
		Handler: p.mux,
	}

	p.tracker = tracker.New(p.cfg)

	// Start Marathon configuration generator
	p.generator = generator.New(p.cfg, p.tracker)
	p.generator.Watch(p.validateConfig)

	p.httpServer.ListenAndServe()
}

func (p *Proxy) getVersion(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, versionResponse, p.cfg.Version)
}

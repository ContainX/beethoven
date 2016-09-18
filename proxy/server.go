package proxy

import (
	"fmt"
	"github.com/ContainX/beethoven/config"
	"github.com/ContainX/beethoven/generator"
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
}

func New(cfg *config.Config) *Proxy {
	return &Proxy{
		cfg: cfg,
	}
}

func (p *Proxy) Serve() {
	mux := mux.NewRouter()
	mux.HandleFunc("/", p.getVersion)

	p.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", p.cfg.HttpPort()),
		Handler: mux,
	}
	p.generator = generator.New(p.cfg)
	p.generator.Watch(p.validateConfig)

	p.httpServer.ListenAndServe()
}

func (p *Proxy) getVersion(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, versionResponse, p.cfg.Version)
}

package proxy

import (
	"fmt"
	"github.com/ContainX/depcon/pkg/encoding"
	"io/ioutil"
	"net/http"
)

func (p *Proxy) getStatus(w http.ResponseWriter, r *http.Request) {
	json, err := encoding.DefaultJSONEncoder().MarshalIndent(p.tracker.GetStatus())
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}
	fmt.Fprint(w, json)
}

func (p *Proxy) getConfig(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadFile(p.cfg.NginxConfig)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err.Error())
	} else {
		w.Write(b)
	}
}

func (p *Proxy) reloadConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if p.cfg.Reload() {
			log.Info("Triggering configuration reload")
			p.generator.ReloadConfiguration()
		}
	} else {
		log.Error("Reload Configuration - invalid method %s", r.Method)
	}
}

func (p *Proxy) reloadAll(w http.ResponseWriter, r *http.Request) {
	if p.cfg.MarathonServiceId != "" {
		// Will trigger reload config on all instances of Beethoven in a cluster
		// if invoked.
	}
}

package proxy

import (
	"fmt"
	"github.com/ContainX/depcon/pkg/encoding"
	"io/ioutil"
	"net/http"
	"strings"
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

// Will trigger reload config on all instances of Beethoven in a cluster
// if invoked.
func (p *Proxy) reloadAll(w http.ResponseWriter, r *http.Request) {
	if p.cfg.MarathonServiceId != "" {
		if r.Method != http.MethodPost {
			log.Error("Reload Configuration - invalid method %s", r.Method)
			return
		}
		if app, err := p.generator.MarathonClient().GetApplication(p.cfg.MarathonServiceId); err == nil {
			// Iterate through all tasks and invoke Reload
			for _, task := range app.Tasks {
				if len(task.Ports) > 0 {
					log.Info("Sending reload to instance: %s:%d", task.Host, task.Ports[0])
					uri := fmt.Sprintf("http://%s:%d/bt/reload/", task.Host, task.Ports[0])
					r, err := http.DefaultClient.Post(uri, "application/json", strings.NewReader("{}"))
					if err != nil {
						log.Error(err.Error())
					} else {
						log.Info("%s:%d response: %d", task.Host, task.Ports[0], r.StatusCode)
					}
				}
			}
		} else {
			log.Error("Error - reload all: %s", err.Error())
		}
	}
}

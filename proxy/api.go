package proxy

import (
	"fmt"
	"github.com/ContainX/depcon/pkg/encoding"
	"net/http"
	"io/ioutil"
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

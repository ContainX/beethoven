package proxy

import (
	"fmt"
	"github.com/ContainX/depcon/pkg/encoding"
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

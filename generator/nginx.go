package generator

import (
	"fmt"
	"github.com/aymerick/raymond"
)

// writeConfiguration writes a temporary nginx configuration based on
// the specified template.  It then validates the configuration for any
// errors
func (g *Generator) writeConfiguration() error {
	tpl, err := raymond.ParseFile(g.cfg.Template)
	if err != nil {
		return fmt.Errorf("Error loading template: %s", err.Error())
	}
	result, err := tpl.Exec(g.templateData.Apps)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Info(result)
	return nil
}

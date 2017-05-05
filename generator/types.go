package generator

import (
	"github.com/ContainX/beethoven/scheduler"
)

type TemplateData struct {
	Apps map[string]*scheduler.App
	Data map[string]interface{}
}

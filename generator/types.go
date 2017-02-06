package generator

import (
	"github.com/ContainX/depcon/marathon"
)

type Task struct {
	Host         string
	Ports        []int
	ServicePorts []int
	StagedAt     string
	StartedAt    string
	Version      string
}

type App struct {
	AppId  string
	Tasks  []Task
	Labels map[string]string
	Env    map[string]string
}

type TemplateData struct {
	Apps map[string]*App
	Data map[string]interface{}
}

func marathonTaskToTask(mt *marathon.Task) Task {
	task := Task{}
	task.Host = mt.Host
	task.Ports = mt.Ports
	task.ServicePorts = mt.ServicePorts
	task.StagedAt = mt.StagedAt
	task.StartedAt = mt.StagedAt
	task.Version = mt.Version
	return task
}

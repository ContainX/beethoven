package scheduler

type App struct {
	AppId  string
	Tasks  []Task
	Labels map[string]string
	Env    map[string]string
}

type Task struct {
	Host         string
	Ports        []int
	ServicePorts []int
	StagedAt     string
	StartedAt    string
	Version      string
}

type ShutdownChan chan bool
type ReloadChan chan bool

type Scheduler interface {
	// Watch for changes using streams and make callbacks to the specified
	// handler when apps have been added, removed or health changes.
	Watch(handler func(proxyConf string))
	// Shutdown the current stream watching
	Shutdown()
	// Fetch all applications/services from the scheduler source
	FetchApps() (map[string]*App, error)
}

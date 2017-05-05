package scheduler

import (
	"github.com/docker/go-connections/nat"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/moby/moby/api/types/swarm"
	"net"
	"time"
)

const (
	// SwarmWatchTime is the default interval when pulling for changes
	SwarmWatchTimeSec = 10
)

type swarmService struct {
	*schedulerService
	started       bool
	client        *docker.Client
	services      []serviceData
	shutdown      ShutdownChan
	watchInterval time.Duration
}

type serviceData struct {
	ServiceName     string
	Name            string
	Labels          map[string]string
	NetworkSettings networkSettings
	Health          string
}

type networkSettings struct {
	Mode     string
	Ports    nat.PortMap
	Networks map[string]*networkData
}

type networkData struct {
	Name       string
	Address    string
	Port       int
	TargetPort int
	Protocol   string
	ID         string
}

func createSwarmScheduler(ss *schedulerService) Scheduler {
	scheduler := &swarmService{schedulerService: ss}
	client, err := docker.NewClient(ss.cfg.Swarm.Endpoint)
	if err != nil {
		panic(err)
	}
	scheduler.client = client
	scheduler.shutdown = make(ShutdownChan, 2)

	if ss.cfg.Swarm.WatchIntervalSecs > 0 {
		scheduler.watchInterval = time.Duration(ss.cfg.Swarm.WatchIntervalSecs) * time.Second
	} else {
		scheduler.watchInterval = SwarmWatchTimeSec * time.Second
	}

	return scheduler
}

// Watch for changes using polling and make callbacks to the specified
// handler when apps have been added, removed or health changes.
// Currently docker doesn't support Swarm events (https://github.com/moby/moby/issues/23827)
func (s *swarmService) Watch(reload chan bool) {
	log.Info("Starting swarm watch")
	s.reload = reload
	ticker := time.NewTicker(s.watchInterval)
	go func(ticker *time.Ticker, s *swarmService) {
		for {
			select {
			case <-ticker.C:
				services, err := s.getServices()
				if err != nil {
					log.Error("Error fetching services from Swarm: %s", err.Error())
				} else {
					previous := s.services
					s.services = services

					// TODO: better comparison until #23827 is implemented
					if len(previous) != len(services) {
						s.reload <- true
					}
				}
			case <-s.shutdown:
				ticker.Stop()
				return
			}
		}
	}(ticker, s)
	s.reload <- true
}

func (s *swarmService) Shutdown() {
	s.shutdown <- true
}

// Fetch all applications/services from the scheduler source
func (s *swarmService) FetchApps() (map[string]*App, error) {
	var err error
	if s.services == nil || len(s.services) == 0 {
		s.services, err = s.getServices()
	}
	converted := s.convertServiceToApp(s.services)
	return converted, err
}

func (s *swarmService) FetchBeethovenInstances() ([]*BeethovenInstance, error) {
	return []*BeethovenInstance{}, nil
}

func (s *swarmService) getServices() ([]serviceData, error) {
	services, err := s.client.ListServices(docker.ListServicesOptions{})
	if err != nil {
		return []serviceData{}, err
	}

	networks, err := s.client.FilteredListNetworks(docker.NetworkFilterOpts{"driver": {"overlay": true}})
	if err != nil {
		log.Debug("Failed to get networks from swarm, error: %s", err)
		return []serviceData{}, err
	}

	networkMap := make(map[string]*docker.Network)

	for _, network := range networks {
		n := network
		networkMap[network.ID] = &n
	}

	serviceDataList := []serviceData{}

	for _, service := range services {
		sdata := parseService(service, networkMap)
		serviceDataList = append(serviceDataList, sdata)
	}

	return serviceDataList, err
}

func parseService(service swarm.Service, networkMap map[string]*docker.Network) serviceData {
	sdata := serviceData{
		ServiceName:     service.Spec.Annotations.Name,
		Name:            service.Spec.Annotations.Name,
		Labels:          service.Spec.Annotations.Labels,
		NetworkSettings: networkSettings{},
	}

	if service.Spec.EndpointSpec != nil {
		if service.Spec.EndpointSpec.Mode == swarm.ResolutionModeVIP {
			sdata.NetworkSettings.Networks = make(map[string]*networkData)
			for _, vip := range service.Endpoint.VirtualIPs {
				networkService := networkMap[vip.NetworkID]
				if networkService != nil {
					ip, _, _ := net.ParseCIDR(vip.Addr)
					network := &networkData{
						Name:       networkService.Name,
						ID:         vip.NetworkID,
						Address:    ip.String(),
						Port:       int(service.Endpoint.Ports[0].PublishedPort),
						TargetPort: int(service.Endpoint.Ports[0].TargetPort),
					}
					sdata.NetworkSettings.Networks[network.Name] = network
				} else {
					log.Debug("Network not found, id: %s", vip.NetworkID)
				}

			}
		}
	}
	return sdata
}

func (s *swarmService) getNetwork(service serviceData) *networkData {
	network := service.NetworkSettings.Networks[s.cfg.Swarm.Network]
	if network != nil {
		return network
	} else {
		log.Warning("Could not find network: %s for service '%s'.  Make sure service is in the Beethoven network'", s.cfg.Swarm.Network, service.Name)
	}

	// Try Swarm ingress network
	network = service.NetworkSettings.Networks["ingress"]
	if network != nil {
		return network
	}

	//for _, network := range service.NetworkSettings.Networks {
	//		return network
	//}
	return nil
}

func (s *swarmService) convertServiceToApp(serviceData []serviceData) map[string]*App {
	apps := make(map[string]*App)
	for _, service := range serviceData {
		network := s.getNetwork(service)
		if network == nil {
			log.Error("Could not find network for: %S, skipping in template", service.Name)
			continue
		}

		swarmTask := Task{}
		swarmTask.Host = network.Address
		swarmTask.Ports = []int{network.TargetPort}
		swarmTask.ServicePorts = []int{network.Port}

		app := App{}
		app.AppId = service.ServiceName
		app.Labels = service.Labels
		app.Tasks = []Task{swarmTask}
		apps[service.ServiceName] = &app
	}
	return apps
}

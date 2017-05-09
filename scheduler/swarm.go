package scheduler

import (
	"errors"
	"github.com/ContainX/beethoven/config"
	"github.com/docker/go-connections/nat"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/moby/moby/api/types/swarm"
	"net"
	"sort"
	"strings"
	"sync"
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
	services      Services
	shutdown      ShutdownChan
	watchInterval time.Duration
	nodes         *swarmNodes
}

type Services []serviceData

type serviceData struct {
	ServiceName     string
	Name            string
	Labels          map[string]string
	NetworkSettings networkSettings
	Health          string
	Port            int
	TargetPort      int
}

type networkSettings struct {
	Mode     string
	Ports    nat.PortMap
	Networks map[string]*networkData
}

type networkData struct {
	Name     string
	Address  string
	Protocol string
	ID       string
}

type swarmNodes struct {
	sync.RWMutex
	nextNode     int
	healthyNodes []nodeData
}

type nodeData struct {
	Id   string
	Addr string
}

func createSwarmScheduler(ss *schedulerService) Scheduler {
	scheduler := &swarmService{schedulerService: ss}
	client, err := newDockerClient(ss.cfg.Swarm)
	if err != nil {
		panic(err)
	}
	scheduler.client = client
	scheduler.shutdown = make(ShutdownChan, 2)
	scheduler.nodes = &swarmNodes{
		healthyNodes: []nodeData{},
	}
	scheduler.updateNodeState()

	if ss.cfg.Swarm.WatchIntervalSecs > 0 {
		scheduler.watchInterval = time.Duration(ss.cfg.Swarm.WatchIntervalSecs) * time.Second
	} else {
		scheduler.watchInterval = SwarmWatchTimeSec * time.Second
	}

	return scheduler
}

func newDockerClient(cfg *config.SwarmConfig) (*docker.Client, error) {
	if strings.HasPrefix(cfg.Endpoint, "unix") {
		return docker.NewClient(cfg.Endpoint)
	} else if cfg.TLSVerify || tlsEnabled(cfg.TLSCert, cfg.TLSCACert, cfg.TLSKey) {
		if cfg.TLSVerify {
			if e, err := pathExists(cfg.TLSCACert); !e || err != nil {
				return nil, errors.New("TLS verification was requested, but CA cert does not exist")
			}
		}
		return docker.NewTLSClient(cfg.Endpoint, cfg.TLSCert, cfg.TLSKey, cfg.TLSCACert)
	}
	return docker.NewClient(cfg.Endpoint)
}

// Watch for changes using polling and make callbacks to the specified
// handler when apps have been added, removed or health changes.
// Currently docker doesn't support Swarm events (https://github.com/moby/moby/issues/23827)
func (s *swarmService) Watch(reload chan bool) {
	log.Info("Starting Swarm Watch...")
	s.reload = reload
	ticker := time.NewTicker(s.watchInterval)

	go func(ticker *time.Ticker, s *swarmService) {
		for {
			select {
			case <-ticker.C:
				services, err := s.getServices()
				s.updateNodeState()
				if err != nil {
					log.Error("Error fetching services from Swarm: %s", err.Error())
				} else {
					previous := s.services
					s.services = services

					// TODO: better comparison until #23827 is implemented
					if topologyChanged(previous, services) {
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

func (s *swarmService) updateNodeState() {
	if s.cfg.Swarm.RouteToNode == false {
		return
	}

	nodes, err := s.client.ListNodes(docker.ListNodesOptions{})
	if err != nil {
		log.Error("Error getting node list: %s", err.Error())
		return
	}

	healthyNodes := []nodeData{}

	for _, node := range nodes {
		if node.Status.State == swarm.NodeStateReady && node.ManagerStatus == nil {
			log.Info("Adding Node: %s, %s", node.ID, node.Status.Addr)
			healthyNodes = append(healthyNodes, nodeData{Id: node.ID, Addr: node.Status.Addr})
		} else {
			log.Info("Skipping Node: %s, %s (online: %v, manager: %v)", node.ID, node.Status.Addr, node.Status.State == swarm.NodeStateReady, node.ManagerStatus != nil)
		}
	}

	s.nodes.Lock()
	s.nodes.healthyNodes = healthyNodes
	s.nodes.Unlock()
}

func (s *swarmService) getServices() (Services, error) {
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

	serviceDataList := Services{}

	for _, service := range services {
		sdata := parseService(service, networkMap)
		serviceDataList = append(serviceDataList, sdata)
	}

	sort.Sort(serviceDataList)

	return serviceDataList, err
}

func parseService(service swarm.Service, networkMap map[string]*docker.Network) serviceData {
	sdata := serviceData{
		ServiceName:     service.Spec.Annotations.Name,
		Name:            service.Spec.Annotations.Name,
		Labels:          service.Spec.Annotations.Labels,
		NetworkSettings: networkSettings{},
	}

	if service.Endpoint.Ports != nil && len(service.Endpoint.Ports) > 0 {
		sdata.Port = int(service.Endpoint.Ports[0].PublishedPort)
		sdata.TargetPort = int(service.Endpoint.Ports[0].TargetPort)
	}

	if service.Spec.EndpointSpec != nil {
		if service.Spec.EndpointSpec.Mode == swarm.ResolutionModeVIP {
			sdata.NetworkSettings.Networks = make(map[string]*networkData)
			for _, vip := range service.Endpoint.VirtualIPs {
				networkService := networkMap[vip.NetworkID]
				if networkService != nil {
					ip, _, _ := net.ParseCIDR(vip.Addr)
					network := &networkData{
						Name:    networkService.Name,
						ID:      vip.NetworkID,
						Address: ip.String(),
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

func (s *swarmService) getAddress(service serviceData) string {
	if s.cfg.Swarm.RouteToNode {
		s.nodes.RLock()
		defer s.nodes.RUnlock()
		return s.nodes.nextNodeAddress()
	}

	network := service.NetworkSettings.Networks[s.cfg.Swarm.Network]
	if network != nil {
		return network.Address
	} else {
		log.Warning("Could not find network: %s for service '%s'.  Make sure service is in the Beethoven network'", s.cfg.Swarm.Network, service.Name)
	}

	// Try Swarm ingress network
	network = service.NetworkSettings.Networks["ingress"]
	if network != nil {
		return network.Address
	}

	for _, network := range service.NetworkSettings.Networks {
		return network.Address
	}
	return ""
}

func (n *swarmNodes) nextNodeAddress() string {
	n.nextNode = n.nextNode + 1
	if len(n.healthyNodes) <= n.nextNode {
		n.nextNode = 0
	}
	nodeData := n.healthyNodes[n.nextNode]
	return nodeData.Addr
}

func topologyChanged(a, b []serviceData) bool {
	if a == nil && b == nil {
		return false
	}

	if a == nil || b == nil {
		return true
	}

	if len(a) != len(b) {
		return true
	}

	for i := range a {
		if a[i].Equal(b[i]) == false {
			return true
		}
	}

	return false
}

func (a serviceData) Equal(b serviceData) bool {
	if a.Name != b.Name {
		return false
	}

	if a.Health != b.Health {
		return false
	}

	if a.TargetPort != b.TargetPort {
		return false
	}

	if a.Port != b.Port {
		return false
	}
	return true
}

func (s Services) Len() int {
	return len(s)
}

func (s Services) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s Services) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s *swarmService) convertServiceToApp(serviceData []serviceData) map[string]*App {
	apps := make(map[string]*App)
	for _, service := range serviceData {
		address := s.getAddress(service)
		if address == "" {
			log.Error("Could not find network address for: %S, skipping in template", service.Name)
			continue
		}

		swarmTask := Task{}
		swarmTask.Host = address
		swarmTask.Ports = []int{service.TargetPort}
		swarmTask.ServicePorts = []int{service.Port}

		app := App{}
		app.AppId = service.ServiceName
		app.Labels = service.Labels
		app.Tasks = []Task{swarmTask}
		apps[service.ServiceName] = &app
	}
	return apps
}

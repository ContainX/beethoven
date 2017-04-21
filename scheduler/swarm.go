package scheduler

type swarmService struct {
	*schedulerService
}

// Watch for changes using streams and make callbacks to the specified
// handler when apps have been added, removed or health changes.
func (s *swarmService) Watch(handler func(proxyConf string)) {

}

// Shutdown the current stream watching
func (s *swarmService) Shutdown() {
	return nil
}

// Fetch all applications/services from the scheduler source
func (s *swarmService) FetchApps() map[string]*App {
	return map[string]*App{}
}


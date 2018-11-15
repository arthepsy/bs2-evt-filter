// +build !windows

package svc

func (s *Service) IsInteractive() (bool, error) {
	return true, nil
}

func (s *Service) Install() error { return nil }
func (s *Service) Remove() error  { return nil }

func (s *Service) Status() string {
	return "not installed"
}

func (s *Service) Start() error    { return nil }
func (s *Service) Stop() error     { return nil }
func (s *Service) Pause() error    { return nil }
func (s *Service) Continue() error { return nil }

func (s *Service) Run(isDebug bool) {
	if s.app != nil {
		s.app()
		s.Stop()
	}
}

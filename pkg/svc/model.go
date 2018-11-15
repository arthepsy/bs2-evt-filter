package svc

type Service struct {
	name    string
	display string
	app     func()
	stop    func()
	args    []string
}

func NewService(name string, display string, app func(), stop func(), args ...string) *Service {
	return &Service{
		name:    name,
		display: display,
		app:     app,
		stop:    stop,
		args:    args,
	}
}

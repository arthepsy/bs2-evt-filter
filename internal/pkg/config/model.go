package config

import (
	"sync"
	"time"
)

type Config struct {
	path       string
	name       string
	lock       *sync.Mutex
	OnReload   func()
	lastReload time.Time

	Service ServiceConf
	Server  ServerConf
	Clients ClientsConf
	Remotes map[string]RemoteConf
}

type ServiceConf struct {
	Name    string
	Display string
}

type ServerConf struct {
	Port int
}

type ClientsConf struct {
	Auth map[string]string
}

type RemoteConf struct {
	BioStar2 BioStar2Conf
	Retry    RetryConf
	Filter   FilterConf
}

type BioStar2Conf struct {
	Url      string
	Username string
	Password string
}

type RetryConf struct {
	Http      int
	WebSocket int
	Session   int
}

type FilterConf struct {
	EventTypeCodes map[string]string
	DeviceIDs      map[string]string
}

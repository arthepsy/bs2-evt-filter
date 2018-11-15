package config

import (
	"log"
	"strings"
	"time"

	"bs2-evt-filter/pkg/sstr"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const (
	defaultRetryHttp      = 15
	defaultRetryWebSocket = 15
	defaultRetrySession   = 10 * 60
)

func NewConfig(path string, name string) *Config {
	return &Config{
		path:       path,
		name:       name,
		OnReload:   nil,
		lastReload: time.Now().Add(time.Second * -5),
	}
}

func (c *Config) Read() {
	viper.SetConfigName(c.name)
	viper.AddConfigPath(c.path)
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("[config] file changed:", e.Name)
		c.reload()
	})
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("[config] error reading: %s\n", err)
	}
	c.reload()
}

func (c *Config) readMap(name string) map[string]string {
	m := make(map[string]string)
	for k, v := range viper.GetStringMapString(name) {
		if len(strings.TrimSpace(v)) == 0 {
			continue
		}
		m[v] = k
	}
	return m

}

func (c *Config) reload() {
	if time.Now().Sub(c.lastReload) < time.Second*1 {
		return
	}
	c.lastReload = time.Now()

	svc := new(ServiceConf)
	svc.Name = viper.GetString("service.name")
	svc.Display = viper.GetString("service.display")
	if len(svc.Display) == 0 {
		svc.Display = svc.Name
	}
	c.Service = *svc

	srv := new(ServerConf)
	srv.Port = viper.GetInt("server.port")
	c.Server = *srv

	clients := new(ClientsConf)
	auth := make(map[string]string)
	for k, v := range viper.GetStringMapString("clients") {
		auth[v] = k
	}
	clients.Auth = auth
	c.Clients = *clients

	remotes := make(map[string]RemoteConf)
	for name := range viper.AllSettings() {
		switch name {
		case "service":
			fallthrough
		case "server":
			fallthrough
		case "clients":
			continue
		}
		remote := new(RemoteConf)

		bs2c := new(BioStar2Conf)
		bs2c.Url = strings.TrimSpace(viper.GetString(name + ".biostar2.url"))
		bs2c.Username = strings.TrimSpace(viper.GetString(name + ".biostar2.username"))
		password := strings.TrimSpace(viper.GetString(name + ".biostar2.password"))
		if sstr.IsProtected(password) {
			np, err := sstr.UnprotectString(strings.ToLower(password))
			if err == nil {
				password = np
			}
		}
		bs2c.Password = password
		if len(bs2c.Url) == 0 || len(bs2c.Username) == 0 || len(bs2c.Password) == 0 {
			continue
		}
		remote.BioStar2 = *bs2c

		retry := new(RetryConf)
		retry.Http = viper.GetInt(name + ".retry.http")
		retry.WebSocket = viper.GetInt(name + ".retry.websocket")
		retry.Session = viper.GetInt(name + ".retry.session")
		if retry.Http <= 0 {
			retry.Http = defaultRetryHttp
		}
		if retry.WebSocket <= 0 {
			retry.WebSocket = defaultRetryWebSocket
		}
		if retry.Session <= 0 {
			retry.Session = defaultRetrySession
		}
		remote.Retry = *retry

		filter := new(FilterConf)
		filter.EventTypeCodes = c.readMap(name + ".filter.event_type_code")
		filter.DeviceIDs = c.readMap(name + ".filter.device_id")
		remote.Filter = *filter

		remotes[name] = *remote
	}
	c.Remotes = remotes

	if c.OnReload != nil {
		c.OnReload()
	}

}

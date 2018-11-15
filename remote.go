package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"bs2-evt-filter/internal/pkg/config"
	"bs2-evt-filter/internal/pkg/ws"
	"bs2-evt-filter/pkg/biostar2"
)

type Remote struct {
	name     string
	hub      *ws.Hub
	config   config.RemoteConf
	bioStar2 *biostar2.API

	recv chan []byte
	send chan []byte
	done chan struct{}
	wait *sync.WaitGroup
	lock *sync.Mutex

	session      chan string
	querySession chan bool
	renewSession *time.Ticker
}

var (
	remotes map[string]*Remote
)

func startRemotes(conf *config.Config, hub *ws.Hub) {
	remotes = make(map[string]*Remote)
	for name, rc := range conf.Remotes {
		r := newRemote(name, rc, hub)
		remotes[name] = r
		go r.start()
	}
}

func reloadRemotes(conf *config.Config, hub *ws.Hub) {
	for name, r := range remotes {
		nrc, ok := conf.Remotes[name]
		if ok {
			oc := r.config.BioStar2
			nc := nrc.BioStar2
			if oc.Url != nc.Url ||
				oc.Username != nc.Username ||
				oc.Password != nc.Password {
				r.stop()
				nr := newRemote(name, nrc, hub)
				remotes[name] = nr
				go nr.start()
			}
			r.lock.Lock()
			r.config.Filter = nrc.Filter
			r.lock.Unlock()
		} else {
			r.stop()
			delete(remotes, name)
		}
	}
	for name, rc := range conf.Remotes {
		_, ok := remotes[name]
		if !ok {
			r := newRemote(name, rc, hub)
			remotes[name] = r
			go r.start()
		}
	}
}

func stopRemotes(conf *config.Config, hub *ws.Hub) {
	for _, r := range remotes {
		r.stop()
	}
}

func newRemote(name string, rc config.RemoteConf, hub *ws.Hub) *Remote {
	return &Remote{
		name:         name,
		hub:          hub,
		config:       rc,
		bioStar2:     nil,
		recv:         make(chan []byte),
		send:         make(chan []byte),
		done:         make(chan struct{}),
		wait:         new(sync.WaitGroup),
		lock:         &sync.Mutex{},
		session:      make(chan string),
		querySession: make(chan bool),
		renewSession: time.NewTicker(time.Duration(rc.Retry.Session) * time.Second),
	}
}

func (r *Remote) log(f string, v ...interface{}) {
	log.Printf("[remote."+r.name+"] "+f, v...)
}

func (r *Remote) start() {
	r.wait.Add(2)
	defer r.wait.Done()
	go r.loop()
	for {
		bs2c := r.config.BioStar2
		r.bioStar2 = biostar2.NewAPI(bs2c.Url, bs2c.Username, bs2c.Password)
		r.log("connecting to websocket")
		r.querySession <- true
		r.bioStar2.WebSocket(r.recv, r.send, r.done)
		select {
		case <-r.done:
			r.log("stopping main routine")
			return
		default:
		}
		retry := time.Duration(r.config.Retry.WebSocket) * time.Second
		r.log("retry connecting to websocket in %v", retry)
		time.Sleep(retry)
	}
}

func (r *Remote) stop() {
	r.log("stopping")
	close(r.done)
	r.wait.Wait()
	r.log("stopped")
}

func (r *Remote) loop() {
	defer r.wait.Done()
	for {
		select {
		case <-r.renewSession.C:
			r.log("renew session")
			go func() { r.querySession <- true }()
		case <-r.querySession:
			r.log("query session")
			go func() {
				for {
					bs2api := r.bioStar2
					r.log("authentication")
					if bs2api.Auth() {
						r.session <- bs2api.SessionID()
						bs2api.StartEvents()
						break
					}
					retry := time.Duration(r.config.Retry.Http) * time.Second
					r.log("retry authentication in %v", retry)
					time.Sleep(retry)
				}
			}()
		case sid := <-r.session:
			r.log("update session")
			msg := []byte(fmt.Sprintf("bs-session-id=%s", sid))
			go func() { r.send <- msg }()
		case msg := <-r.recv:
			resp, ok := biostar2.ParseResponse(msg)
			if ok {
				r.log("response code: %s, msg: %s", resp.Code, resp.Message)
				if resp.Code != "0" {
					retry := time.Duration(r.config.Retry.WebSocket) * time.Second
					r.log("invalid auth. rety in %v", retry)
					go func() {
						time.Sleep(retry)
						r.querySession <- true
					}()
				}
				continue
			}
			e, ok := biostar2.ParseEvent(msg)
			if ok {
				r.lock.Lock()
				filter := r.config.Filter
				ok := filterEvent(&filter, e)
				r.lock.Unlock()
				if ok {
					r.log("filtered event: %v", e)
					r.hub.Broadcast(msg)
				}
			}
		case <-r.done:
			r.log("stopping loop routine")
			return
		}
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"bs2-evt-filter/internal/pkg/config"
	"bs2-evt-filter/internal/pkg/ws"
)

type Server struct {
	conf *config.Config
	hub  *ws.Hub
	srv  *http.Server
	done chan struct{}
	wait *sync.WaitGroup
}

func newServer(conf *config.Config, hub *ws.Hub) *Server {
	return &Server{
		conf: conf,
		hub:  hub,
		srv:  nil,
		done: make(chan struct{}),
		wait: new(sync.WaitGroup)}
}

func (s *Server) log(f string, v ...interface{}) {
	log.Printf("[srv] "+f, v...)
}

func (s *Server) start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "BioStar2 Event Filter\n")
	})
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		s.hub.Client(w, r)
	})
	s.srv = &http.Server{}
	s.wait.Add(1)
	go func() {
		defer s.wait.Done()
		for {
			addr := fmt.Sprintf(":%d", s.conf.Server.Port)
			s.log("starting on %s", addr)
			*s.srv = http.Server{Addr: addr, Handler: mux}
			err := s.srv.ListenAndServeTLS(appPath("server.crt"), appPath("server.key"))
			if err != http.ErrServerClosed {
				s.log("serve error: %v", err)
			}
			retry := 10 * time.Second
			select {
			case <-s.done:
				return
			default:
				s.log("retry in %v", retry)
			}
			timeout := time.After(retry)
			select {
			case <-s.done:
				return
			case <-timeout:
				break
			}
		}
	}()
}

func (s *Server) shutdown() {
	if err := s.srv.Shutdown(context.Background()); err != nil {
		s.log("shutdown error: %v", err)
	}
	s.wait.Wait()
	s.log("shutdown")

}
func (s *Server) reload() {
	addr := fmt.Sprintf(":%d", s.conf.Server.Port)
	if s.srv.Addr != addr {
		s.log("restarting")
		s.shutdown()
	}

}

func (s *Server) stop() {
	s.log("stopping")
	close(s.done)
	s.shutdown()
	s.log("stopped")
}

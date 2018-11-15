// +build windows

package svc

import (
	"fmt"
	//"log"
	//"os"
	//"path/filepath"
	//"strings"
	"sync"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

func (s *Service) IsInteractive() (bool, error) {
	return svc.IsAnInteractiveSession()
}

func (s *Service) Install() error {
	appPath := s.appPath()
	// todo: empty = error
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	svc, err := m.OpenService(s.name)
	if err == nil {
		svc.Close()
		return fmt.Errorf("service %s already exists", s.name)
	}
	svc, err = m.CreateService(s.name, appPath, mgr.Config{
		DisplayName: s.display,
		StartType:   mgr.StartAutomatic,
	}, s.args...)
	if err != nil {
		return err
	}
	defer svc.Close()
	err = eventlog.InstallAsEventCreate(s.name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		svc.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}
	return nil
}

func (s *Service) Remove() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	svc, err := m.OpenService(s.name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", s.name)
	}
	defer svc.Close()
	err = svc.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(s.name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	return nil
}

func (s *Service) Status() string {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Sprintf("mgr error: %v", err)
	}
	defer m.Disconnect()
	wsvc, err := m.OpenService(s.name)
	if err != nil {
		return "not installed"
	}
	defer wsvc.Close()
	status, err := wsvc.Query()
	if err != nil {
		return fmt.Sprintf("query error: %v", err)
	}
	switch status.State {
	case svc.Stopped:
		return "stopped"
	case svc.StartPending:
		return "starting"
	case svc.StopPending:
		return "stopping"
	case svc.Running:
		return "running"
	case svc.ContinuePending:
		return "continuing"
	case svc.PausePending:
		return "pausing"
	case svc.Paused:
		return "paused"
	default:
		return fmt.Sprintf("unknown: %d", status.State)
	}
}

func (s *Service) Start() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	svc, err := m.OpenService(s.name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer svc.Close()
	err = svc.Start(s.args...)
	if err != nil {
		return fmt.Errorf("could not start service: %v", err)
	}
	return nil
}

func (s *Service) controlService(c svc.Cmd, to svc.State) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	svc, err := m.OpenService(s.name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer svc.Close()
	status, err := svc.Control(c)
	if err != nil {
		return fmt.Errorf("could not send control=%d: %v", c, err)
	}
	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = svc.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}
	return nil
}

func (s *Service) Stop() error {
	return s.controlService(svc.Stop, svc.Stopped)
}

func (s *Service) Pause() error {
	return s.controlService(svc.Pause, svc.Paused)
}

func (s *Service) Continue() error {
	return s.controlService(svc.Continue, svc.Running)
}

func (s *Service) Run(isDebug bool) {
	var err error
	var elog debug.Log
	if isDebug {
		elog = debug.New(s.name)
	} else {
		elog, err = eventlog.Open(s.name)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", s.name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(s.name, s)
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", s.name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", s.name))
}

func (s *Service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	//s.log.Info(1, strings.Join(args, "-"))
	wait := new(sync.WaitGroup)
	wait.Add(1)
	if s.app != nil {
		go func() {
			defer wait.Done()
			s.app()
			s.Stop()
		}()
	}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				if s.stop != nil {
					s.stop()
				}
				wait.Wait()
				return false, 0
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				continue loop
				// s.log.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
}

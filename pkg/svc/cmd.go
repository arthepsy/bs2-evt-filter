package svc

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

func (s *Service) usage(errmsg string) {
	prog := path.Base(s.appPath())
	if len(s.args) > 0 {
		prog = prog + " " + strings.Join(s.args, " ")
	}
	fmt.Fprintf(os.Stderr, "%s\n"+
		"\nUsage:\n"+
		"\t%s <command>\n"+
		"\nThe commands are:\n"+
		"\tinstall     install service\n"+
		"\tremove      remove service\n"+
		"\tdebug       run in debug mode\n"+
		"\tstatus      print service status\n"+
		"\tstart       start service\n"+
		"\tstop        stop service\n"+
		"\tpause       pause service\n"+
		"\tcontinue    continue service\n"+
		"\n",
		errmsg, prog)
	os.Exit(2)
}

func (s *Service) Cmd() {
	isIntSess, err := s.IsInteractive()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}
	if !isIntSess {
		s.Run(false)
		return
	}
	pos := len(s.args) + 1
	if len(os.Args) < (pos + 1) {
		s.usage("no command specified")
	}

	cmd := strings.ToLower(os.Args[pos])
	switch cmd {
	case "debug":
		s.Run(true)
		return
	case "install":
		err = s.Install()
	case "remove":
		err = s.Remove()
	case "status":
		fmt.Fprintln(os.Stdout, s.Status())
		return
	case "start":
		err = s.Start()
	case "stop":
		err = s.Stop()
	case "pause":
		err = s.Pause()
	case "continue":
		err = s.Continue()
	default:
		s.usage(fmt.Sprintf("unknown command: %s", cmd))
	}
	if err != nil {
		log.Fatalf("failed to %s %s: %v", cmd, s.name, err)
	}
	return
}

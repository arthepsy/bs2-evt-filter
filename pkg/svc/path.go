package svc

import (
	"os"
	"path/filepath"
)

func (s *Service) appPath() string {
	prog, err := os.Executable()
	if err != nil {
		prog = os.Args[0]
	}
	prog, err = filepath.Abs(prog)
	if err != nil {
		prog = ""
	}
	return prog
}

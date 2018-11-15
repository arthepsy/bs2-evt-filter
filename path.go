package main

import (
	"os"
	"path"
	"path/filepath"
)

var AppDir = getAppDir()

func getAppDir() string {
	prog, err := os.Executable()
	if err != nil {
		prog = os.Args[0]
	}
	prog, err = filepath.Abs(prog)
	if err != nil {
		prog = "."
	}
	return filepath.Dir(prog)
}

func appPath(elem ...string) string {
	if len(elem) == 0 {
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
	elem = append([]string{AppDir}, elem...)
	return path.Join(elem...)
}

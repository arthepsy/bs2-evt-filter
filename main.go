package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"bs2-evt-filter/internal/pkg/config"
	"bs2-evt-filter/pkg/sstr"
	"bs2-evt-filter/pkg/svc"
)

func usage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\n"+
		"\t%s <command>\n"+
		"\nThe commands are:\n"+
		"\tstart        start program\n"+
		"\tservice      service commands\n"+
		"\tprotect      protect string\n"+
		"\tunprotect    unprotect string\n"+
		"\n", path.Base(appPath()))
	os.Exit(2)
}

func protectUsage(cmd string) {
	fmt.Fprintf(os.Stderr, "\nUsage:\n\t%s <string>\n\n", cmd)
	os.Exit(2)
}

func getApp() *App {
	c := config.NewConfig(AppDir, "config")
	app := newApp(c)
	c.Read()
	c.OnReload = app.reload
	return app
}

func errOut(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "start":
		getApp().run()
	case "service":
		app := getApp()
		svcConf := app.config.Service
		if len(svcConf.Name) == 0 {
			fmt.Fprintln(os.Stderr, "service.name not defined")
			os.Exit(1)
		}
		svc := svc.NewService(svcConf.Name, svcConf.Display, app.run, app.stop, cmd)
		svc.Cmd()
	case "protect":
		if len(os.Args) != 3 {
			protectUsage(cmd)
		}
		b, err := sstr.ProtectString(os.Args[2])
		errOut(err)
		fmt.Fprintln(os.Stdout, b)
	case "unprotect":
		if len(os.Args) != 3 {
			protectUsage(cmd)
		}
		s, err := sstr.UnprotectString(os.Args[2])
		errOut(err)
		fmt.Fprintln(os.Stdout, s)
	default:
		usage()
	}
}

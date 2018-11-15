package main

import (
	"io"
	"log"
	"os"
	"os/signal"

	"bs2-evt-filter/internal/pkg/config"
	"bs2-evt-filter/internal/pkg/ws"
)

type App struct {
	config  *config.Config
	reloadC chan bool
	stopC   chan os.Signal
}

func newApp(config *config.Config) *App {
	return &App{
		config:  config,
		reloadC: make(chan bool),
		stopC:   make(chan os.Signal, 1),
	}
}

func (app *App) reload() {
	go func() { app.reloadC <- true }()
}

func (app *App) stop() {
	go func() { app.stopC <- os.Interrupt }()
}

func (app *App) run() {
	signal.Notify(app.stopC, os.Interrupt)

	logFile, err := os.OpenFile(appPath("bs2-evt-filter.log"), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	mw := io.MultiWriter(logFile, os.Stdout)
	log.SetOutput(mw)

	hub := ws.NewHub()
	hub.UpdateAuth(app.config.Clients.Auth)

	go hub.Run()
	server := newServer(app.config, hub)
	server.start()
	go startRemotes(app.config, hub)

	log.Println("[main] starting")
	for {
		select {
		case <-app.reloadC:
			log.Println("[main] reloading")
			hub.UpdateAuth(app.config.Clients.Auth)
			server.reload()
			reloadRemotes(app.config, hub)
		case <-app.stopC:
			log.Println("[main] stopping")
			server.stop()
			stopRemotes(app.config, hub)
			log.Println("[main] finished")
			return
		}
	}
}

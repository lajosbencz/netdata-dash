package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/gammazero/nexus/v3/router"
)

const (
	Address = "localhost:11000"
	Realm   = "meta"
)

func main() {
	cfg := router.Config{
		RealmConfigs: []*router.RealmConfig{
			{
				URI:           Realm,
				AnonymousAuth: true,
				AllowDisclose: true,
			},
		},
	}
	nxr, _ := router.NewRouter(&cfg, log.Default())
	wss := router.NewWebsocketServer(nxr)

	closer, _ := wss.ListenAndServe(Address)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
	closer.Close()
}

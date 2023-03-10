package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/lajosbencz/netdata-dash/app"
	"github.com/lajosbencz/netdata-dash/router"
)

func main() {
	var (
		realm  = "netdata"
		wsHost = "0.0.0.0"
		wsPort = 16666
	)
	flag.StringVar(&realm, "realm", realm, "realm")
	flag.StringVar(&wsHost, "ws-host", wsHost, "websocket host")
	flag.IntVar(&wsPort, "ws-port", wsPort, "websocket port")
	flag.Parse()

	rtr, err := router.NewRouter(realm, wsHost, wsPort, "127.0.0.1", 0)
	if err != nil {
		log.Fatalln(err)
	}
	httpRouter := http.NewServeMux()
	httpRouter.Handle("/ws/", rtr)
	httpRouter.Handle("/", http.FileServer(http.Dir("./web/")))
	go func() {
		http.ListenAndServe(rtr.HttpListenAddress, httpRouter)
	}()

	cli, err := client.ConnectLocal(rtr.Router, client.Config{
		Realm: realm,
	})
	if err != nil {
		log.Fatalln(err)
	}

	if !cli.HasFeature("broker", wamp.FeatureSubMetaAPI) {
		log.Fatal("Broker does not have", wamp.FeatureSubMetaAPI, "feature")
	}
	if !cli.HasFeature("broker", wamp.FeatureSessionMetaAPI) {
		log.Fatal("Broker does not have", wamp.FeatureSessionMetaAPI, "feature")
	}
	if !cli.HasFeature("dealer", wamp.FeatureSessionMetaAPI) {
		log.Fatal("Dealer does not have", wamp.FeatureSessionMetaAPI, "feature")
	}

	defer cli.Close()

	myApp := app.NewApp(cli)
	myApp.RunLoop()
	myApp.Close()
	rtr.Close()
	fmt.Println("fin.")
}

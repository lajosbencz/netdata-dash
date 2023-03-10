package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gammazero/nexus/v3/client"
	"log"
	"os"
	"os/signal"

	"github.com/lajosbencz/netdata-dash/agent"
)

const (
	ConfigFileName = "config.json"
)

func main() {
	var (
		paramHostName = ""
	)
	flag.StringVar(&paramHostName, "hostname", paramHostName, "Overwrite registration hostname (useful for debugging)")
	flag.Parse()

	agentConfig := agent.DefaultConfig()
	if err := agentConfig.FromFile(ConfigFileName); err != nil && !os.IsNotExist(err) {
		log.Fatalln(err)
	}
	fmt.Printf("%#v\n", agentConfig)

	ctxRun, cancelRun := context.WithCancel(context.Background())

	wampClient, err := client.ConnectNet(ctxRun, fmt.Sprintf("http://%s/ws/", agentConfig.Router.Format()), client.Config{
		Realm:         agentConfig.Realm,
		Debug:         true,
		Serialization: client.MSGPACK,
	})
	if err != nil {
		log.Fatalln(err)
	}
	a, err := agent.NewAgent("localhost", agentConfig, wampClient)
	if err != nil {
		log.Fatalln(err)
	}
	if err := a.Watch("system.cpu", "system.ram"); err != nil {
		log.Fatalln(err)
	}
	go a.Run(ctxRun)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown

	cancelRun()
	fmt.Println("fin.")
}

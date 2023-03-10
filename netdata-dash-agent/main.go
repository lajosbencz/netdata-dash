package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/lajosbencz/netdata-dash/agent"
)

func main() {
	a, err := agent.NewAgent(agent.DefaultConfig())
	if err != nil {
		log.Fatalln(err)
	}
	a.Watch("system.cpu", "system.ram")
	ctxRun, cancelRun := context.WithCancel(context.Background())
	go a.Run(ctxRun)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown

	cancelRun()
	fmt.Println("fin.")
}

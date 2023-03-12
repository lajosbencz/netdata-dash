package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lajosbencz/netdata-dash/app"
)

func main() {
	var (
		realm     = "netdata"
		host      = "0.0.0.0"
		port      = 16666
		agentName = ""
	)
	flag.StringVar(&realm, "realm", realm, "realm")
	flag.StringVar(&host, "host", host, "network host")
	flag.IntVar(&port, "port", port, "network port")
	flag.StringVar(&agentName, "agent", agentName, "run local agent with provided hostname")
	flag.Parse()

	wampRouter, wampClient, wampServer, err := newWamp(realm)
	if err != nil {
		log.Fatalln(err)
	}
	defer wampClient.Close()
	defer wampRouter.Close()

	httpRouter := http.NewServeMux()

	httpRouter.Handle("/ws/", wampServer)
	httpRouter.HandleFunc("/", newHttp())

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	listenAddress := fmt.Sprintf("%s:%d", host, port)

	failed := make(chan error, 1)

	httpServer := &http.Server{Addr: listenAddress, Handler: httpRouter}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			failed <- err
		}
	}()

	myApp := app.NewApp(wampClient)
	go func() {
		if err := myApp.RunLoop(); err != nil {
			failed <- err
		}
	}()

	log.Printf("listening on http://%s\n", listenAddress)

	runCount := 2
out:
	for {
		select {
		case err := <-failed:
			log.Println(err)
			runCount--
			if runCount == 0 {
				log.Println("all workers failed, exiting...")
				break out
			}
		case <-shutdown:
			break out
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	httpServer.Shutdown(ctx)
	cancel()
}

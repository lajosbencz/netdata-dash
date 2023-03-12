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

	"github.com/lajosbencz/netdata-dash/pkg/app"
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

	wampRouter, wampClient, wampServer, err := newWamp(realm, log.Default())
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

	httpsListener, err := newTlsListener(listenAddress, "localhost", "dev", "dev", "dev", 3600)
	if err != nil {
		log.Fatalln(err)
	}
	httpsServer := &http.Server{Addr: listenAddress, Handler: httpRouter}
	go func() {
		if err := httpsServer.Serve(httpsListener); err != nil {
			failed <- err
		}
	}()

	myApp := app.NewApp(wampClient)
	go func() {
		if err := myApp.RunLoop(shutdown); err != nil {
			failed <- err
		}
	}()

	log.Printf("listening on https://%s\n", listenAddress)

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
	httpsServer.Shutdown(ctx)
	cancel()
}

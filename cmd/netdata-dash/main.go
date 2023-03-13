package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lajosbencz/netdata-dash/pkg/app"
)

const (
	DefaultConfigPath = "dash.json"
)

func main() {
	var (
		realm = "netdata"
		host  = "0.0.0.0"
		port  = 16666
	)
	flag.StringVar(&realm, "realm", realm, "realm")
	flag.StringVar(&host, "host", host, "network host")
	flag.IntVar(&port, "port", port, "network port")
	flag.Parse()

	wampRouter, wampClient, wampServer, err := newWamp(realm, log.Default())
	if err != nil {
		log.Fatalln(err)
	}
	defer wampClient.Close()
	defer wampRouter.Close()

	httpRouter := http.NewServeMux()

	httpRouter.Handle("/ws/", wampServer)
	muxIndex, err := newHttp()
	if err != nil {
		log.Fatalln(err)
	}
	httpRouter.Handle("/", muxIndex)

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

	la := listenAddress
	if strings.HasPrefix(la, "0.0.0.0:") {
		la = "localhost:" + strings.TrimPrefix(la, "0.0.0.0:")
	}
	log.Printf("listening on https://%s\n", la)

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

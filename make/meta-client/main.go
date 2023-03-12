package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
)

const (
	Address = "localhost:11000"
	Realm   = "meta"
	Topic   = "chart.data._.lazos-nb._.system.cpu"
)

func main() {
	ctx := context.Background()
	c, _ := client.ConnectNet(ctx, "http://"+Address, client.Config{
		Realm: Realm,
	})

	c.Subscribe(Topic, func(event *wamp.Event) {
		fmt.Printf("chart.data %#v\n", event)
	}, nil)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
	c.Close()
}

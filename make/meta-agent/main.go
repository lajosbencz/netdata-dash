package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

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
	c, err := client.ConnectNet(ctx, "http://"+Address, client.Config{
		Realm: Realm,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ticker := time.NewTicker(3 * time.Second)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

out:
	for {
		select {
		case <-ticker.C:
			go func() {
				res, err := c.Call(ctx, string(wamp.MetaProcSubLookup), nil, wamp.List{Topic}, nil, nil)
				if err != nil {
					log.Fatalln(err)
				}
				if len(res.Arguments) != 0 {
					if subID, _ := wamp.AsID(res.Arguments[0]); subID > 0 {
						fmt.Printf("lookup: %v\n", subID)
					}
				}
			}()
			go func() {
				res, err := c.Call(ctx, string(wamp.MetaProcSubList), nil, nil, nil, nil)
				if err != nil {
					log.Fatalln(err)
				}
				if len(res.Arguments) != 0 {
					if lists, ok := wamp.AsDict(res.Arguments[0]); ok {
						if exact, ok := wamp.AsList(lists["exact"]); ok {
							for _, subID := range exact {
								res, err := c.Call(ctx, string(wamp.MetaProcSubGet), nil, wamp.List{subID}, nil, nil)
								if err != nil {
									log.Fatalln(err)
								}
								if len(res.Arguments) != 0 {
									if subInfo, ok := wamp.AsDict(res.Arguments[0]); ok {
										if topic, ok := subInfo["uri"]; ok {
											log.Printf("sub: %s\n", topic.(string))
										}
									}
								}
							}
						}
					}
				}
			}()
		case <-shutdown:
			break out
		}
	}

	c.Close()
}

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/lajosbencz/netdata-dash/pkg/agent"
	"github.com/lajosbencz/netdata-dash/pkg/core"
)

const (
	defaultConfigPath = "agent.json"
)

func main() {
	var (
		authUser, authPassword string
	)
	verboseOutput := false
	configPath := defaultConfigPath
	agentConfig := agent.DefaultConfig()

	flag.StringVar(&authUser, "u", core.AgentRole, "Auth user")
	flag.StringVar(&authPassword, "p", "", "Auth password")
	flag.StringVar(&configPath, "config", configPath, "Path of config file")
	flag.StringVar(&agentConfig.HostName, "hostname", agentConfig.HostName, "Overwrite registration hostname (useful for debugging)")
	flag.Var(&agentConfig.HostTags, "tags", "Comma separated list of host_tags")
	flag.StringVar(&agentConfig.Realm, "realm", agentConfig.Realm, "Realm")
	flag.StringVar(&agentConfig.Dash.Host, "dash-host", agentConfig.Dash.Host, "Netdata Dash host")
	flag.IntVar(&agentConfig.Dash.Port, "dash-port", agentConfig.Dash.Port, "Netdata Dash port")
	flag.StringVar(&agentConfig.Netdata.Host, "netdata-host", agentConfig.Netdata.Host, "Netdata host")
	flag.IntVar(&agentConfig.Netdata.Port, "netdata-port", agentConfig.Netdata.Port, "Netdata port")
	flag.BoolVar(&verboseOutput, "vv", verboseOutput, "Verbose output")
	flag.Parse()

	if authPassword == "" {
		log.Fatalln("no auth credentials provided (-p)")
	}

	if err := agentConfig.FromFile(configPath); err != nil && (!os.IsNotExist(err) || configPath != defaultConfigPath) {
		log.Fatalln(err)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	go func() {
		for {
			wampClient, err := agent.NewClient(context.Background(), agentConfig)
			if err != nil {
				log.Println(err)
			} else {
				agentLogger := log.Default()
				a, err := agent.NewAgent(agentConfig.HostName, agentConfig, wampClient, agentLogger)
				if err != nil {
					log.Println(err)
				} else {
					if err := a.Run(); err != nil {
						log.Println(err)
					}
					if err := wampClient.Close(); err != nil {
						log.Println(err)
					}
				}
			}

			time.Sleep(time.Second * 5)
			log.Println("retrying connection...")
		}
	}()
	<-shutdown
}

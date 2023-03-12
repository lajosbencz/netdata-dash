package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"

	"github.com/lajosbencz/netdata-dash/pkg/agent"
)

const (
	defaultConfigPath = "config.json"
)

func main() {
	verboseOutput := false
	configPath := defaultConfigPath
	agentConfig := agent.DefaultConfig()
	if osHostname, err := os.Hostname(); err == nil {
		agentConfig.HostName = osHostname
	}
	flag.StringVar(&configPath, "config", configPath, "Path of config.json")
	flag.StringVar(&agentConfig.HostName, "hostname", agentConfig.HostName, "Overwrite registration hostname (useful for debugging)")
	flag.Var(&agentConfig.HostTags, "tags", "Comma separated list of host_tags")
	flag.StringVar(&agentConfig.Realm, "realm", agentConfig.Realm, "Realm")
	flag.StringVar(&agentConfig.Dash.Host, "dash-host", agentConfig.Dash.Host, "Netdata Dash host")
	flag.IntVar(&agentConfig.Dash.Port, "dash-port", agentConfig.Dash.Port, "Netdata Dash port")
	flag.StringVar(&agentConfig.Netdata.Host, "netdata-host", agentConfig.Netdata.Host, "Netdata host")
	flag.IntVar(&agentConfig.Netdata.Port, "netdata-port", agentConfig.Netdata.Port, "Netdata port")
	flag.BoolVar(&verboseOutput, "vv", verboseOutput, "Verbose output")
	flag.Parse()

	if err := agentConfig.FromFile(configPath); err != nil && (!os.IsNotExist(err) || configPath != defaultConfigPath) {
		log.Fatalln(err)
	}

	wampUrl := fmt.Sprintf("http://%s/ws/", agentConfig.Dash.Format())
	wampConfig := client.Config{
		Realm:         agentConfig.Realm,
		Debug:         verboseOutput,
		Logger:        log.Default(),
		Serialization: client.MSGPACK,
		HelloDetails: wamp.Dict{
			agent.HostnameKey: agentConfig.HostName,
		},
	}
	wampClient, err := client.ConnectNet(context.Background(), wampUrl, wampConfig)
	if err != nil {
		log.Fatalln(err)
	}
	defer wampClient.Close()

	agentLogger := log.Default()
	a, err := agent.NewAgent(agentConfig.HostName, agentConfig, wampClient, agentLogger)
	if err != nil {
		log.Fatalln(err)
	}

	if err := a.Run(); err != nil {
		log.Println(err)
	}
}

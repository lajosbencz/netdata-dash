package agent

import (
	"context"
	"fmt"
	"github.com/gammazero/nexus/v3/client"
	"log"
	"time"

	"github.com/lajosbencz/netdata-dash/netdata"
	"github.com/lajosbencz/netdata-dash/utils"
)

type Agent struct {
	HostName    string
	Config      *Config
	metricsList utils.StringsUnique
	chartsData  netdata.ApiDataCharts
	wampClient  *client.Client
}

func NewAgent(hostName string, config *Config, wampClient *client.Client) (*Agent, error) {
	if config == nil {
		config = DefaultConfig()
	}
	a := &Agent{
		HostName:    hostName,
		Config:      config,
		metricsList: utils.StringsUnique{},
		chartsData:  netdata.ApiDataCharts{},
		wampClient:  wampClient,
	}
	return a, nil
}

func (r *Agent) RouterAddress() string {
	return fmt.Sprintf("%s:%d", r.Config.Router.Host, r.Config.Router.Port)
}

func (r *Agent) NetdataAddress() string {
	return fmt.Sprintf("%s:%d", r.Config.Netdata.Host, r.Config.Netdata.Port)
}

func (r *Agent) UpdateCharts() error {
	charts, err := netdata.ApiCharts(r.NetdataAddress())
	if err != nil {
		return err
	}
	r.chartsData = *charts
	return nil
}

func (r *Agent) Watch(metric ...string) error {
	r.metricsList.Add(metric...)
	return nil
}

func (r *Agent) Unwatch(metric ...string) error {
	r.metricsList.Remove(metric...)
	return nil
}

func (r *Agent) Run(ctx context.Context) {
	if err := r.UpdateCharts(); err != nil {
		log.Fatalln(err)
	}
	ticker := time.NewTicker(time.Duration(r.chartsData.UpdateEvery) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Agent stopped via context")
			return
		case <-ticker.C:
			// fmt.Printf("%#v\n", r.chartsData)
		}
	}
}

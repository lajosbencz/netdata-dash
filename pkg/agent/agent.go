package agent

import (
	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/stdlog"
	"github.com/gammazero/nexus/v3/wamp"

	"github.com/lajosbencz/netdata-dash/pkg/netdata"
	"github.com/lajosbencz/netdata-dash/pkg/utils"
)

const (
	HostnameKey = "hostname"
)

type Agent struct {
	HostName    string
	Config      *Config
	metricsList utils.StringsUnique
	chartsData  netdata.HostData
	wampClient  *client.Client
	topicIds    map[wamp.ID]string
	logger      stdlog.StdLog
}

func NewAgent(hostName string, config *Config, wampClient *client.Client, logger stdlog.StdLog) (*Agent, error) {
	if config == nil {
		config = DefaultConfig()
	}
	a := &Agent{
		HostName:    hostName,
		Config:      config,
		metricsList: utils.StringsUnique{},
		chartsData:  netdata.HostData{},
		wampClient:  wampClient,
		topicIds:    map[wamp.ID]string{},
		logger:      logger,
	}
	return a, nil
}

func (r *Agent) Watch(metrics ...string) int {
	n := 0
	for _, m := range metrics {
		if r.metricsList.Add(m) != 0 {
			n++
			r.logger.Printf("+%s\n", m)
		}
	}
	return n
}

func (r *Agent) Unwatch(metrics ...string) int {
	n := 0
	for _, m := range metrics {
		if r.metricsList.Remove(m) != 0 {
			n++
			r.logger.Printf("-%s\n", m)
		}
	}
	return n
}

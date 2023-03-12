package agent

import (
	"fmt"
	"log"
	"strings"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"

	"github.com/lajosbencz/netdata-dash/core"
	"github.com/lajosbencz/netdata-dash/netdata"
	"github.com/lajosbencz/netdata-dash/utils"
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
	logger      *log.Logger
}

func NewAgent(hostName string, config *Config, wampClient *client.Client, logger *log.Logger) (*Agent, error) {
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

func (r *Agent) RouterAddress() string {
	return fmt.Sprintf("%s:%d", r.Config.Dash.Host, r.Config.Dash.Port)
}

func (r *Agent) NetdataAddress() string {
	return fmt.Sprintf("%s:%d", r.Config.Netdata.Host, r.Config.Netdata.Port)
}

func (r *Agent) Watch(metrics ...string) int {
	n := 0
	for _, m := range metrics {
		if r.metricsList.Add(m) != 0 {
			n++
			r.logger.Printf("watching metric: %s\n", m)
		}
	}
	return n
}

func (r *Agent) Unwatch(metrics ...string) int {
	n := 0
	for _, m := range metrics {
		if r.metricsList.Remove(m) != 0 {
			n++
			r.logger.Printf("unwatching metric: %s\n", m)
		}
	}
	return n
}

func (r *Agent) onSubCreate(event *wamp.Event) {
	if len(event.Arguments) > 1 {
		if details, ok := wamp.AsDict(event.Arguments[1]); ok {
			subID, _ := wamp.AsID(details["id"])
			topic, _ := wamp.AsString(details["uri"])
			if strings.HasPrefix(topic, core.TopicChartDataHostPrefix(r.HostName)) {
				parts := strings.SplitN(topic, core.TopicPartDelimiter, 3)
				var metricName string = parts[2]
				r.topicIds[subID] = metricName
				_ = r.Watch(metricName)
			}
		}
	}
}

func (r *Agent) onSubDelete(event *wamp.Event) {
	if len(event.Arguments) > 1 {
		if subID, ok := wamp.AsID(event.Arguments[1]); ok {
			if metricName, ok := r.topicIds[subID]; ok {
				r.Unwatch(metricName)
				delete(r.topicIds, subID)
			}
		}
	}
}

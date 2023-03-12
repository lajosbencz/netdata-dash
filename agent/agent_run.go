package agent

import (
	"os"
	"os/signal"
	"time"

	"github.com/gammazero/nexus/v3/wamp"
	"github.com/lajosbencz/netdata-dash/core"
	"github.com/lajosbencz/netdata-dash/netdata"
)

func (r *Agent) Run() error {
	if err := r.UpdateCharts(); err != nil {
		return err
	}
	if err := r.UpdateSubscribedMetrics(); err != nil {
		return err
	}
	r.wampClient.Subscribe(string(wamp.MetaEventSubOnCreate), r.onSubCreate, nil)
	r.wampClient.Subscribe(string(wamp.MetaEventSubOnDelete), r.onSubDelete, nil)
	ticker := time.NewTicker(time.Duration(r.chartsData.UpdateEvery) * time.Second)
	defer ticker.Stop()
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	r.logger.Printf("found %d metrics, updating every %ds\n", len(r.chartsData.Charts), r.chartsData.UpdateEvery)
	for {
		select {
		case <-shutdown:
			return nil
		case <-ticker.C:
			if len(r.metricsList) != 0 {
				data, err := netdata.ApiMetrics(r.Config.Netdata.Format(), r.metricsList)
				if err != nil {
					return err
				}
				for metricName, metricData := range *data {
					topic := core.TopicChartData(r.HostName, metricName)
					kwArgs := wamp.Dict{"metricData": metricData, "metricName": metricName}
					if err := r.wampClient.Publish(topic, nil, nil, kwArgs); err != nil {
						return err
					}
				}
				r.logger.Printf("published %d metrics\n", len(r.metricsList))
			}
			if err := r.wampClient.Publish(core.TopicHostHeartbeat(r.HostName), nil, nil, nil); err != nil {
				return err
			}
		}
	}
}

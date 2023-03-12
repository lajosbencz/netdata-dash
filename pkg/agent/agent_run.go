package agent

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/lajosbencz/netdata-dash/pkg/core"
	"github.com/lajosbencz/netdata-dash/pkg/netdata"
)

func (r *Agent) Run() error {
	if err := r.UpdateCharts(); err != nil {
		return err
	}
	if err := r.UpdateSubscribedMetrics(); err != nil {
		return err
	}
	if err := r.wampClient.Register(core.ProcedureChartData(r.HostName), r.rpcChartData, wamp.Dict{wamp.OptDiscloseCaller: true}); err != nil {
		return err
	}
	if err := r.wampClient.Subscribe(string(wamp.MetaEventSubOnCreate), r.onSubCreate, nil); err != nil {
		return err
	}
	if err := r.wampClient.Subscribe(string(wamp.MetaEventSubOnDelete), r.onSubDelete, nil); err != nil {
		return err
	}

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
				r.logger.Printf("<%d\n", len(r.metricsList))
			}
			if err := r.wampClient.Publish(core.TopicHostHeartbeat(r.HostName), nil, nil, nil); err != nil {
				return err
			}
		}
	}
}

func (r *Agent) rpcChartData(ctx context.Context, inv *wamp.Invocation) client.InvokeResult {
	if metric, ok := wamp.AsString(inv.ArgumentsKw["metric"]); ok {
		after := wamp.OptionInt64(inv.ArgumentsKw, "after")
		before := wamp.OptionInt64(inv.ArgumentsKw, "before")
		if after == before {
			after = before - 600
		}
		apiData, err := netdata.ApiChartData(r.Config.Netdata.Format(), metric, after, before)
		if err != nil {
			return client.InvokeResult{Err: "app.netdata.api.error", Args: wamp.List{err.Error()}}
		}
		inter, err := json.Marshal(apiData)
		if err != nil {
			return client.InvokeResult{Err: "app.netdata.json.error", Args: wamp.List{err.Error()}}
		}
		kwData := wamp.Dict{}
		if err := json.Unmarshal(inter, &kwData); err != nil {
			return client.InvokeResult{Err: "app.netdata.json.error", Args: wamp.List{err.Error()}}
		}
		return client.InvokeResult{Kwargs: kwData}
	}
	return client.InvokeResult{Err: wamp.ErrInvalidArgument}
}

func (r *Agent) onSubCreate(event *wamp.Event) {
	if len(event.Arguments) > 1 {
		if details, ok := wamp.AsDict(event.Arguments[1]); ok {
			subID, _ := wamp.AsID(details["id"])
			topic, _ := wamp.AsString(details["uri"])
			if strings.HasPrefix(topic, core.TopicChartDataHostPrefix(r.HostName)) {
				parts := strings.SplitN(topic, core.TopicPartDelimiter, 3)
				metricName := parts[2]
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

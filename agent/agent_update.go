package agent

import (
	"context"
	"strings"

	"github.com/gammazero/nexus/v3/wamp"
	"github.com/lajosbencz/netdata-dash/core"
	"github.com/lajosbencz/netdata-dash/netdata"
	"github.com/lajosbencz/netdata-dash/utils"
)

func (r *Agent) UpdateCharts() error {
	charts, err := netdata.ApiCharts(r.NetdataAddress())
	if err != nil {
		return err
	}
	r.chartsData = *charts
	return nil
}

func (r *Agent) UpdateSubscribedMetrics() error {
	res, err := r.wampClient.Call(context.Background(), string(wamp.MetaProcSubList), nil, nil, nil, nil)
	if err != nil {
		return err
	}
	if len(res.Arguments) != 0 {
		if lists, ok := wamp.AsDict(res.Arguments[0]); ok {
			if exact, ok := wamp.AsList(lists["exact"]); ok {
				// we got sensible data, clear existing metrics
				r.metricsList = utils.StringsUnique{}
				for _, subID := range exact {
					res, err := r.wampClient.Call(context.Background(), string(wamp.MetaProcSubGet), nil, wamp.List{subID}, nil, nil)
					if err != nil {
						return err
					}
					if len(res.Arguments) != 0 {
						if subInfo, ok := wamp.AsDict(res.Arguments[0]); ok {
							if topicRaw, ok := subInfo["uri"]; ok {
								topic, ok := topicRaw.(string)
								if ok && strings.HasPrefix(topic, core.TopicChartDataHostPrefix(r.HostName)) {
									parts := strings.SplitN(topic, core.TopicPartDelimiter, 3)
									if len(parts) == 3 {
										r.Watch(parts[2])
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
}

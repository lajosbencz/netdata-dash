package core

import (
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/lajosbencz/netdata-dash/pkg/utils"
)

type MetricList = utils.StringsUnique

type HostMetrics map[string]MetricList

func (r *HostMetrics) Add(hostName string, metricNames MetricList) {
	_, ok := (*r)[hostName]
	if !ok {
		(*r)[hostName] = MetricList{}
	}
	for _, m := range metricNames {
		mt := (*r)[hostName]
		mt.Add(m)
	}
}

func (r *HostMetrics) Has(hostName string) bool {
	_, ok := (*r)[hostName]
	return ok
}

func (r *HostMetrics) HasMetric(metricName string) bool {
	for _, m := range *r {
		if m.Has(metricName) {
			return true
		}
	}
	return false
}

type ClientHostMetrics map[wamp.ID]HostMetrics

func (r *ClientHostMetrics) Has(clientId wamp.ID) bool {
	_, ok := (*r)[clientId]
	return ok
}

func (r *ClientHostMetrics) HasHost(hostName string) bool {
	for _, h := range *r {
		if h.Has(hostName) {
			return true
		}
	}
	return false
}

func (r *ClientHostMetrics) HasMetric(metricName string) bool {
	for _, m := range *r {
		if m.HasMetric(metricName) {
			return true
		}
	}
	return false
}

func (r *ClientHostMetrics) GetAllHosts() []string {
	l := utils.StringsUnique{}
	for _, clientRegs := range *r {
		for hostName := range clientRegs {
			l.Add(hostName)
		}
	}
	return l
}

func (r *ClientHostMetrics) GetAllMetrics() MetricList {
	l := MetricList{}
	for _, clientRegs := range *r {
		for _, metrics := range clientRegs {
			for _, m := range metrics {
				l.Add(m)
			}
		}
	}
	return l
}

func (r *ClientHostMetrics) GetHostMetrics(hostName string) MetricList {
	l := MetricList{}
	for _, clientRegs := range *r {
		m, ok := clientRegs[hostName]
		if ok {
			l.Add(m...)
		}
	}
	return l
}

func (r *ClientHostMetrics) Register(clientId wamp.ID, hostName string, metricName ...string) {
	if !(*r).Has(clientId) {
		(*r)[clientId] = HostMetrics{}
	}
	clientRegs := (*r)[clientId]
	if !clientRegs.Has(hostName) {
		clientRegs[hostName] = MetricList{}
	}
	clientMetrics := clientRegs[hostName]
	clientMetrics.Add(metricName...)
}

func (r *ClientHostMetrics) Unregister(clientId wamp.ID) {
	delete(*r, clientId)
}

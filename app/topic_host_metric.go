package app

import (
	"strings"
	"sync"

	"github.com/lajosbencz/netdata-dash/utils"
)

type topicHostMetric struct {
	safeMap sync.Map
}

func topicHostMetricKey(hostName, metricName string) string {
	return hostName + ";" + metricName
}

func (r *topicHostMetric) Hash(hostName string, metricName string) string {
	key := topicHostMetricKey(hostName, metricName)
	hash := utils.HashMD5(key)
	if _, has := r.safeMap.Load(hash); !has {
		r.safeMap.Store(hash, key)
	}
	return hash
}

func (r *topicHostMetric) Unhash(hash string) (string, string, bool) {
	v, has := r.safeMap.Load(hash)
	if !has {
		return "", "", false
	}
	parts := strings.SplitN(v.(string), ";", 2)
	return parts[0], parts[1], true
}

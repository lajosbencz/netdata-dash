package core_test

import (
	"testing"

	"github.com/lajosbencz/netdata-dash/core"
)

func TestConstants(t *testing.T) {
	if s := core.TopicChartData("a", "b"); s != "chart.data._.a._.b" {
		t.Error("TopicChartData", s)
	}
	if s := core.TopicChartDataHostPrefix("a"); s != "chart.data._.a._." {
		t.Error("TopicChartDataHostPrefix", s)
	}
	if s := core.TopicHostHeartbeat("a"); s != "host.heartbeat._.a" {
		t.Error("TopicHostHeartbeat", s)
	}
}

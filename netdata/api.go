package netdata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/lajosbencz/netdata-dash/core"
)

const (
	urlCharts     = "http://%s/api/v1/charts"
	urlChartData  = "http://%s/api/v1/data?chart=%s&after=%d&before=%d"
	urlAllMetrics = "http://%s/api/v1/allmetrics?format=%s&filter=%s"
)

func ApiCharts(hostName string) (*ChartHost, error) {
	d := &ChartHost{}
	url := fmt.Sprintf(urlCharts, hostName)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&d); err != nil {
		return nil, err
	}
	return d, nil
}

func ApiChartData(hostName, metricName string, after, before int64) (*ChartData, error) {
	if after >= before {
		after = before - 60
	}
	url := fmt.Sprintf(urlChartData, hostName, metricName, after, before)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	d := &ChartData{}
	if err := json.NewDecoder(res.Body).Decode(&d); err != nil {
		return nil, err
	}
	return d, nil
}

func ApiMetrics(hostName string, metrics core.MetricList) (*AllMetrics, error) {
	filter := strings.Join(metrics, "%20")
	url := fmt.Sprintf(urlAllMetrics, hostName, "json", filter)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	d := &AllMetrics{}
	if err := json.NewDecoder(res.Body).Decode(&d); err != nil {
		return nil, err
	}
	return d, nil
}

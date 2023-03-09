package netdata

import (
	"github.com/gammazero/nexus/v3/wamp"
)

type Dimension struct {
	Name  string   `json:"name"`
	Value *float64 `json:"value,omitempty"`
}

type Host struct {
	Hostname string `json:"hostname"`
}

type Chart struct {
	Id             string                 `json:"id"`
	Name           string                 `json:"name"`
	Type           string                 `json:"type"`
	Family         string                 `json:"family"`
	Context        string                 `json:"context"`
	Title          string                 `json:"title"`
	Priority       int                    `json:"priority"`
	Plugin         string                 `json:"plugin"`
	Modules        string                 `json:"modules"`
	Units          string                 `json:"units"`
	DataUrl        string                 `json:"data_url"`
	ChartType      string                 `json:"chart_type"`
	Duration       int                    `json:"duration"`
	FirstEntry     int                    `json:"first_entry"`
	LastEntry      int                    `json:"last_entry"`
	UpdateEvery    int                    `json:"update_every"`
	Dimensions     map[string]Dimension   `json:"dimensions"`
	ChartVariables string                 `json:"chart_variables"`
	Green          *string                `json:"green"`
	Red            *string                `json:"red"`
	Alarms         map[string]Alarm       `json:"alarms"`
	ChartLabels    map[string]string      `json:"chart_labels"`
	Functions      map[string]interface{} `json:"functions"`
}

type Alarm struct {
	Id          int    `json:"id"`
	Status      string `json:"status"`
	Units       string `json:"units"`
	UpdateEvery int    `json:"update_every"`
}

type Metric struct {
	Name        string               `json:"name"`
	Family      string               `json:"family"`
	Context     string               `json:"context"`
	Units       string               `json:"units"`
	LastUpdated int                  `json:"last_updated"`
	Dimensions  map[string]Dimension `json:"dimensions"`
}

// ChartHost /api/v1/allmetrics
type AllMetrics map[string]Metric

// ChartHost /api/v1/charts
type ChartHost struct {
	Hostname        string           `json:"hostname"`
	Version         string           `json:"version"`
	ReleaseChannel  string           `json:"release_channel"`
	OS              string           `json:"os"`
	Timezone        string           `json:"timezone"`
	UpdateEvery     int              `json:"update_every"`
	History         int              `json:"history"`
	MemoryMode      string           `json:"memory_mode"`
	CustomInfo      string           `json:"custom_info"`
	Charts          map[string]Chart `json:"charts"`
	ChartsCount     int              `json:"charts_count"`
	DimensionsCount int              `json:"dimensions_count"`
	AlarmsCount     int              `json:"alarms_count"`
	RrdMemoryBytes  int              `json:"rrd_memory_bytes"`
	HostsCount      int              `json:"hosts_count"`
	Hosts           []Host           `json:"hosts"`
}

type ChartDataPoint []interface{}

func (r ChartDataPoint) GetTime() int {
	return r[0].(int)
}

func (r ChartDataPoint) GetValues() []float64 {
	l := []float64{}
	for _, v := range r[1:] {
		l = append(l, v.(float64))
	}
	return l
}

// ChartData /api/v1/data?chart=<metric>
type ChartData struct {
	Labels []string         `json:"labels"`
	Data   []ChartDataPoint `json:"data"`
}

func (r *ChartData) GetTime(index int) int {
	return r.Data[index][0].(int)
}

func ChartDataFromDict(d wamp.Dict) ChartData {
	labels := []string{}
	if labelList, ok := wamp.AsList(d["labels"]); ok {
		for _, v := range labelList {
			labels = append(labels, v.(string))
		}
	}
	data := []ChartDataPoint{}
	if dataList0, ok := wamp.AsList(d["data"]); ok {
		for _, v0 := range dataList0 {
			if dataList1, ok := wamp.AsList(v0); ok {
				list1 := []interface{}{}
				for _, v1 := range dataList1 {
					list1 = append(list1, v1)
				}
				data = append(data, list1)
			}
		}
	}
	return ChartData{
		Labels: labels,
		Data:   data,
	}
}

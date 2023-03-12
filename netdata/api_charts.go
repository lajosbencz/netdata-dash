package netdata

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gammazero/nexus/v3/wamp"
)

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
	ChartVariables map[string]interface{} `json:"chart_variables"`
	Green          *interface{}           `json:"green,omitempty"`
	Red            *interface{}           `json:"red,omitempty"`
	Alarms         map[string]Alarm       `json:"alarms"`
	ChartLabels    map[string]string      `json:"chart_labels"`
	Functions      map[string]interface{} `json:"functions"`
}

type HostData struct {
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

func ApiCharts(hostName string) (*HostData, error) {
	query := url.Values{}
	query.Add("format", "json")
	url := fmt.Sprintf("http://%s/api/v1%s?%s", hostName, apiPathCharts, query.Encode())
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	d := &HostData{}
	if err := json.NewDecoder(res.Body).Decode(&d); err != nil {
		return nil, err
	}
	// log.Println(url)
	// log.Printf("%#v\n", *d)
	return d, nil
}

func ChartFromWampDict(dict wamp.Dict) Chart {
	r := Chart{}
	if d, err := json.Marshal(dict); err == nil {
		if err = json.Unmarshal(d, &r); err != nil {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}
	return r
}

func HostDataFromWampDict(dict wamp.Dict) HostData {
	r := HostData{}
	if d, err := json.Marshal(dict); err == nil {
		if err = json.Unmarshal(d, &r); err != nil {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}
	return r
}

package netdata

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gammazero/nexus/v3/wamp"
	"github.com/lajosbencz/netdata-dash/pkg/core"
)

type Metric struct {
	Name        string               `json:"name"`
	Family      string               `json:"family"`
	Context     string               `json:"context"`
	Units       string               `json:"units"`
	LastUpdated int                  `json:"last_updated"`
	Dimensions  map[string]Dimension `json:"dimensions"`
}

type AllMetrics map[string]Metric

func MetricFromWampDict(dict wamp.Dict) Metric {
	r := Metric{}
	if d, err := json.Marshal(dict); err == nil {
		if err = json.Unmarshal(d, &r); err != nil {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}
	return r
}

func ApiMetrics(hostName string, metrics core.MetricList) (*AllMetrics, error) {
	filter := strings.Join(metrics, " ")
	query := url.Values{}
	query.Add("format", "json")
	query.Add("filter", filter)
	url := fmt.Sprintf("http://%s/api/v1%s?%s", hostName, apiPathAllMetrics, query.Encode())
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

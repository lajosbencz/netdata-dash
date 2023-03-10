package netdata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gammazero/nexus/v3/wamp"
)

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

func ApiChartData(hostName, metricName string, after, before int64) (*ChartData, error) {
	if after >= before {
		after = before - 60
	}
	query := url.Values{}
	query.Add("format", "json")
	query.Add("chart", metricName)
	query.Add("after", fmt.Sprintf("%d", after))
	query.Add("before", fmt.Sprintf("%d", before))
	url := fmt.Sprintf("http://%s/api/v1%s?%s", hostName, apiPathChartData, query.Encode())
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

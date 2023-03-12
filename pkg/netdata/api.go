package netdata

const (
	apiPathCharts     = "/charts"
	apiPathChartData  = "/data"
	apiPathAllMetrics = "/allmetrics"
)

type Dimension struct {
	Name  string   `json:"name"`
	Value *float64 `json:"value,omitempty"`
}

type Alarm struct {
	Id          int    `json:"id"`
	Status      string `json:"status"`
	Units       string `json:"units"`
	UpdateEvery int    `json:"update_every"`
}

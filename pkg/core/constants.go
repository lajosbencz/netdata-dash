package core

const (
	TopicPartDelimiter = "._."

	TopicHostList            = "host.list"
	TopicHostJoin            = "host.join"
	TopicHostLeave           = "host.leave"
	TopicHostHeartbeatPrefix = "host.heartbeat"

	TopicChartDataPrefix = "chart.data"
)

func TopicChartDataHostPrefix(hostName string) string {
	return TopicChartDataPrefix + TopicPartDelimiter + hostName + TopicPartDelimiter
}

func TopicChartData(hostName, metricName string) string {
	return TopicChartDataHostPrefix(hostName) + metricName
}

func TopicHostHeartbeat(hostName string) string {
	return TopicHostHeartbeatPrefix + TopicPartDelimiter + hostName
}

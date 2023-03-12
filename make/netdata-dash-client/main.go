package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/lajosbencz/netdata-dash/core"
	"github.com/lajosbencz/netdata-dash/netdata"
	"github.com/lajosbencz/netdata-dash/utils"
)

func getHostList(wampClient *client.Client) []string {
	rpcHostList, err := wampClient.Call(context.Background(), "host.list", nil, nil, nil, nil)
	if err != nil {
		log.Fatalln(err)
	}
	hostList, _ := wamp.AsList(rpcHostList.ArgumentsKw["list"])
	strList := utils.StringsUnique{}
	for _, v := range hostList {
		if str, ok := wamp.AsString(v); ok {
			strList.Add(str)
		}
	}
	return strList
}

var (
	realm      = "netdata"
	host       = "localhost"
	port       = 16666
	metricList = utils.StringsUnique{}
)

func onMetricData(hostName, metricName string, values netdata.Metric) {
	log.Printf("%s %s data: %v\n", hostName, metricName, values.Dimensions)
}

func subscribeToMetric(wampClient *client.Client, hostName, metricName string) error {
	topic := core.TopicChartData(hostName, metricName)
	err := wampClient.Subscribe(topic, func(event *wamp.Event) {
		if dataIf, ok := event.ArgumentsKw["metricData"]; ok {
			if dataKw, ok := wamp.AsDict(dataIf); ok {
				data := netdata.MetricFromWampDict(dataKw)
				onMetricData(hostName, metricName, data)
			}
		}
	}, nil)
	if err == nil {
		log.Printf("subscribed to %s\n", topic)
	}
	return err
}

func main() {
	flag.StringVar(&realm, "realm", realm, "WAMP realm")
	flag.StringVar(&host, "host", host, "Netdata-Dash Server host")
	flag.IntVar(&port, "port", port, "Netdata-Dash Server port")
	flag.Var(&metricList, "metrics", "Comma separated list of metrics to subscribe to")
	flag.Parse()
	ctx := context.Background()
	wampConfig := client.Config{
		Realm: realm,
	}
	wampUrl := fmt.Sprintf("http://%s:%d/ws/", host, port)
	wampClient, err := client.ConnectNet(ctx, wampUrl, wampConfig)
	if err != nil {
		log.Fatalln(err)
	}

	if len(metricList) == 0 {
		metricList = append(metricList, "system.cpu", "system.ram")
	}
	hostList := getHostList(wampClient)
	log.Printf("initial host list: %#v\n", hostList)
	for _, hostName := range hostList {
		for _, metricName := range metricList {
			if err := subscribeToMetric(wampClient, hostName, metricName); err != nil {
				log.Fatalln(err)
			}
		}
	}

	wampClient.Subscribe("host.join", func(event *wamp.Event) {
		hostName, _ := wamp.AsString(event.ArgumentsKw["hostname"])
		log.Printf("host joined: %s\n", hostName)
		for _, metricName := range metricList {
			go subscribeToMetric(wampClient, hostName, metricName)
		}
	}, nil)

	wampClient.Subscribe("host.leave", func(event *wamp.Event) {
		hostName, _ := wamp.AsString(event.ArgumentsKw["hostname"])
		log.Printf("host left: %s\n", hostName)
		for _, metricName := range metricList {
			topic := core.TopicChartData(hostName, metricName)
			go wampClient.Unsubscribe(topic)
		}
	}, nil)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
}

package app

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/lajosbencz/netdata-dash/core"
	"github.com/lajosbencz/netdata-dash/netdata"
)

type App struct {
	wampClient    *client.Client
	registrations core.ClientHostMetrics
	topics        topicHostMetric
}

func NewApp(wampClient *client.Client) *App {
	return &App{
		wampClient:    wampClient,
		registrations: core.ClientHostMetrics{},
		topics:        topicHostMetric{},
	}
}

func (r *App) Close() error {
	return r.wampClient.Close()
}

func (r *App) onSubscribe(event *wamp.Event) {
	log.Printf("onSubscribe %#v\n", event)
	if len(event.Arguments) != 0 {
		if details, ok := wamp.AsDict(event.Arguments[0]); ok {
			sessionID, _ := wamp.AsID(details["session"])
			authID, _ := wamp.AsString(details["authid"])
			topic, _ := wamp.AsString(details["topic"])
			log.Printf("Client %v subscribed (authid=%s) (topic=%s)\n", sessionID, authID, topic)
			if strings.HasPrefix(topic, "chart.data.") {
				hostName, metricName, ok := r.topics.Unhash(topic[11:])
				if !ok {
					log.Println(errors.New("failed to unhash topic: " + topic))
				}
				r.registrations.Register(sessionID, hostName, metricName)
				log.Printf("client: %v\nhost: %s\nmetric: %s\n", sessionID, hostName, metricName)
			}
		}
	}
}

func (r *App) onUnsubscribe(event *wamp.Event) {
	log.Printf("onUnsubscribe %#v\n", event)
	if len(event.Arguments) != 0 {
		if details, ok := wamp.AsDict(event.Arguments[0]); ok {
			sessionID, _ := wamp.AsID(details["session"])
			authID, _ := wamp.AsString(details["authid"])
			topic, _ := wamp.AsString(details["topic"])
			log.Printf("Client %v unsubscribed (authid=%s) (topic=%s)\n", sessionID, authID, topic)
		}
	}
}

func (r *App) onSessionLeave(event *wamp.Event) {
	log.Printf("onSessionLeave %#v\n", event)
	if len(event.Arguments) != 0 {
		if details, ok := wamp.AsDict(event.Arguments[0]); ok {
			sessionID, _ := wamp.AsID(details["session"])
			authID, _ := wamp.AsString(details["authid"])
			log.Printf("Client %v left the session (authid=%s)\n", sessionID, authID)
		}
	}
}

func (r *App) RunLoop() {
	if err := r.wampClient.Subscribe(string(wamp.MetaEventSubOnSubscribe), (*r).onSubscribe, nil); err != nil {
		log.Fatalln(err)
	}
	if err := r.wampClient.Subscribe(string(wamp.MetaEventSubOnUnsubscribe), (*r).onUnsubscribe, nil); err != nil {
		log.Fatalln(err)
	}
	if err := r.wampClient.Subscribe(string(wamp.MetaEventSessionOnLeave), (*r).onSessionLeave, nil); err != nil {
		log.Fatalln(err)
	}
	r.wampClient.Register("chart.data", r.RpcChartData, wamp.Dict{"disclose_caller": true})
	r.wampClient.Register("chart.topic", func(ctx context.Context, i *wamp.Invocation) client.InvokeResult {
		host, _ := wamp.AsString(i.ArgumentsKw["host"])
		metric, _ := wamp.AsString(i.ArgumentsKw["metric"])
		topic := "chart.data." + r.topics.Hash(host, metric)
		return client.InvokeResult{Kwargs: wamp.Dict{
			"topic": topic,
		}}
	}, wamp.Dict{"disclose_caller": true})

	dataTicker := time.NewTicker(1 * time.Second)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
out:
	for {
		select {
		case <-dataTicker.C:
			for _, hostName := range r.registrations.GetAllHosts() {
				metrics := r.registrations.GetHostMetrics(hostName)
				go func(cli *client.Client, hostName string, metrics core.MetricList) {
					data, err := netdata.ApiMetrics(hostName, metrics)
					if err != nil {
						log.Println(err)
						return
					}
					for _, d := range *data {
						hash := r.topics.Hash(hostName, d.Name)
						topic := "chart.data." + hash
						cli.Publish(topic, wamp.Dict{}, wamp.List{}, wamp.Dict{"data": d})
					}
				}(r.wampClient, hostName, metrics)
			}
		case <-shutdown:
			break out
		}
	}
}

func (r *App) RpcChartData(ctx context.Context, i *wamp.Invocation) client.InvokeResult {
	clientId, _ := wamp.AsID(i.Details["caller"])
	host, _ := wamp.AsString(i.ArgumentsKw["host"])
	metric, _ := wamp.AsString(i.ArgumentsKw["metric"])
	after, _ := wamp.AsInt64(i.ArgumentsKw["after"])
	before, _ := wamp.AsInt64(i.ArgumentsKw["before"])
	if after == before && after == 0 {
		after = -1
	}
	data, err := netdata.ApiChartData(host, metric, after, before)
	if err != nil {
		return client.InvokeResult{Err: wamp.URI("error.app"), Args: wamp.List{err.Error()}}
	}
	info, err := netdata.ApiMetrics(host, core.MetricList{metric})
	if err != nil {
		return client.InvokeResult{Err: wamp.URI("error.app"), Args: wamp.List{err.Error()}}
	}
	r.registrations.Register(clientId, host, metric)
	return client.InvokeResult{
		Kwargs: wamp.Dict{
			"info": info,
			"data": data,
		},
	}
}

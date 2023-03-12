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
	"github.com/lajosbencz/netdata-dash/agent"
	"github.com/lajosbencz/netdata-dash/core"
	"github.com/lajosbencz/netdata-dash/netdata"
	"github.com/lajosbencz/netdata-dash/utils"
)

type App struct {
	wampClient    *client.Client
	agents        map[wamp.ID]string
	registrations core.ClientHostMetrics
	topics        topicHostMetric
}

func NewApp(wampClient *client.Client) *App {
	return &App{
		wampClient:    wampClient,
		agents:        map[wamp.ID]string{},
		registrations: core.ClientHostMetrics{},
		topics:        topicHostMetric{},
	}
}

func (r App) Agents() map[wamp.ID]string {
	return r.agents
}

func (r App) AgentHosts() []string {
	strList := utils.StringsUnique{}
	for _, v := range r.agents {
		if str, ok := wamp.AsString(v); ok {
			strList.Add(str)
		}
	}
	return strList
}

func (r *App) Close() error {
	return r.wampClient.Close()
}

func (r *App) onSessionJoin(event *wamp.Event) {
	if len(event.Arguments) != 0 {
		if details, ok := wamp.AsDict(event.Arguments[0]); ok {
			log.Printf("onSessionJoin %#v\n", details)
			sessionID, _ := wamp.AsID(details["session"])
			if hostName, ok := wamp.AsString(details[agent.HostnameKey]); ok {
				r.agents[sessionID] = hostName
				r.wampClient.Publish("host.join", wamp.Dict{}, wamp.List{}, wamp.Dict{
					agent.HostnameKey: hostName,
				})
				if wampList, ok := wamp.AsList(r.AgentHosts()); ok {
					r.wampClient.Publish("host.list", wamp.Dict{}, wampList, nil)
				}
			}
		}
	}
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
		if sessionID, ok := wamp.AsID(event.Arguments[0]); ok {
			if hostName, in := r.agents[sessionID]; in {
				delete(r.agents, sessionID)
				r.wampClient.Publish("host.leave", wamp.Dict{}, wamp.List{}, wamp.Dict{
					agent.HostnameKey: hostName,
				})
				if wampList, ok := wamp.AsList(r.AgentHosts()); ok {
					r.wampClient.Publish("host.list", wamp.Dict{}, wampList, nil)
				}
			}
		}
	}
}

func (r *App) RunLoop() error {
	if err := r.wampClient.Subscribe(string(wamp.MetaEventSessionOnJoin), r.onSessionJoin, nil); err != nil {
		return err
	}
	if err := r.wampClient.Subscribe(string(wamp.MetaEventSubOnSubscribe), r.onSubscribe, nil); err != nil {
		return err
	}
	if err := r.wampClient.Subscribe(string(wamp.MetaEventSubOnUnsubscribe), r.onUnsubscribe, nil); err != nil {
		return err
	}
	if err := r.wampClient.Subscribe(string(wamp.MetaEventSessionOnLeave), r.onSessionLeave, nil); err != nil {
		return err
	}
	r.wampClient.Register("host.list", r.RpcHosts, wamp.Dict{"disclose_caller": true})
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
	return nil
}

func (r *App) RpcHosts(ctx context.Context, i *wamp.Invocation) client.InvokeResult {
	list := utils.StringsUnique{}
	for _, v := range r.agents {
		list.Add(v)
	}
	return client.InvokeResult{
		Kwargs: wamp.Dict{
			"list": list,
		},
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

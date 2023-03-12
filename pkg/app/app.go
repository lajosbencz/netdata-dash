package app

import (
	"context"
	"os"
	"time"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/lajosbencz/netdata-dash/pkg/agent"
	"github.com/lajosbencz/netdata-dash/pkg/core"
	"github.com/lajosbencz/netdata-dash/pkg/utils"
)

type App struct {
	wampClient *client.Client
	agents     map[wamp.ID]string
}

func NewApp(wampClient *client.Client) *App {
	return &App{
		wampClient: wampClient,
		agents:     map[wamp.ID]string{},
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

func (r *App) onSessionJoin(event *wamp.Event) {
	if len(event.Arguments) != 0 {
		if details, ok := wamp.AsDict(event.Arguments[0]); ok {
			sessionID, _ := wamp.AsID(details["session"])
			if hostName, ok := wamp.AsString(details[agent.HostnameKey]); ok {
				r.agents[sessionID] = hostName
				r.wampClient.Publish(core.TopicHostJoin, wamp.Dict{}, wamp.List{}, wamp.Dict{
					agent.HostnameKey: hostName,
				})
				if wampList, ok := wamp.AsList(r.AgentHosts()); ok {
					r.wampClient.Publish(core.TopicHostList, wamp.Dict{}, wampList, nil)
				}
			}
		}
	}
}

func (r *App) onSessionLeave(event *wamp.Event) {
	if len(event.Arguments) != 0 {
		if sessionID, ok := wamp.AsID(event.Arguments[0]); ok {
			if hostName, in := r.agents[sessionID]; in {
				delete(r.agents, sessionID)
				r.wampClient.Publish(core.TopicHostLeave, wamp.Dict{}, wamp.List{}, wamp.Dict{
					agent.HostnameKey: hostName,
				})
				if wampList, ok := wamp.AsList(r.AgentHosts()); ok {
					r.wampClient.Publish(core.TopicHostList, wamp.Dict{}, wampList, nil)
				}
			}
		}
	}
}

func (r *App) RunLoop(shutdown chan os.Signal) error {
	if err := r.wampClient.Subscribe(string(wamp.MetaEventSessionOnJoin), r.onSessionJoin, nil); err != nil {
		return err
	}
	if err := r.wampClient.Subscribe(string(wamp.MetaEventSessionOnLeave), r.onSessionLeave, nil); err != nil {
		return err
	}
	r.wampClient.Register(core.TopicHostList, r.RpcHosts, wamp.Dict{wamp.OptDiscloseCaller: true})

	dataTicker := time.NewTicker(1 * time.Second)
out:
	for {
		select {
		case <-dataTicker.C:
			// log.Println("tick")
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

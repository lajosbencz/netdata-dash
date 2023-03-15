package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/lajosbencz/netdata-dash/pkg/app"
)

func NewClient(ctx context.Context, agentConfig *Config) (*client.Client, error) {
	wampUrl := fmt.Sprintf("https://%s/ws/", agentConfig.Dash.Format())
	wampConfig := client.Config{
		Realm:         agentConfig.Realm,
		Debug:         agentConfig.Debug,
		Logger:        log.Default(),
		Serialization: client.MSGPACK,
		HelloDetails: wamp.Dict{
			"authid":    agentConfig.User,
			HostnameKey: agentConfig.HostName,
		},
		AuthHandlers: map[string]client.AuthFunc{
			"ticket": func(challenge *wamp.Challenge) (signature string, details wamp.Dict) {
				return agentConfig.Secret, wamp.Dict{}
			},
		},
	}
	return app.NewTlsClient(ctx, wampUrl, wampConfig)
}

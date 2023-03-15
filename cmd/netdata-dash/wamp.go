package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/router"
	"github.com/gammazero/nexus/v3/router/auth"
	"github.com/gammazero/nexus/v3/stdlog"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/lajosbencz/netdata-dash/pkg/core"
)

func newWamp(realm string, logger stdlog.StdLog) (router.Router, *client.Client, *router.WebsocketServer, error) {
	ks, err := WatchFileKeyStore(DefaultConfigPath, logger)
	if err != nil {
		return nil, nil, nil, err
	}
	crAuth := auth.NewTicketAuthenticator(ks, time.Second*5)
	routerConfig := &router.Config{
		Debug: false,
		RealmConfigs: []*router.RealmConfig{
			{
				URI:            wamp.URI(realm),
				AnonymousAuth:  true,
				AllowDisclose:  true,
				Authenticators: []auth.Authenticator{crAuth},
				Authorizer:     &authz{core.AgentRole},
			},
		},
	}
	wampRouter, err := router.NewRouter(routerConfig, nil)
	if err != nil {
		log.Fatalln(err)
	}

	wampConfig := client.Config{
		Debug:  false,
		Realm:  realm,
		Logger: logger,
	}
	wampClient, err := client.ConnectLocal(wampRouter, wampConfig)
	if err != nil {
		return nil, nil, nil, err
	}

	if !wampClient.HasFeature("broker", wamp.FeatureSessionMetaAPI) {
		return nil, nil, nil, fmt.Errorf("broker does not have %s feature", wamp.FeatureSessionMetaAPI)
	}
	if !wampClient.HasFeature("dealer", wamp.FeatureSessionMetaAPI) {
		return nil, nil, nil, fmt.Errorf("fealer does not have %s feature", wamp.FeatureSessionMetaAPI)
	}
	if !wampClient.HasFeature("broker", wamp.FeatureSubMetaAPI) {
		return nil, nil, nil, fmt.Errorf("broker does not have %s feature", wamp.FeatureSubMetaAPI)
	}

	wampServer := router.NewWebsocketServer(wampRouter)
	wampServer.Upgrader.EnableCompression = true
	wampServer.Upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	wampServer.EnableTrackingCookie = true
	wampServer.KeepAlive = 30 * time.Second

	return wampRouter, wampClient, wampServer, nil
}

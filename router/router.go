package router

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gammazero/nexus/v3/router"
	"github.com/gammazero/nexus/v3/wamp"
)

type Router struct {
	router.Router
	Realm             string
	wsCloser          io.Closer
	tcpCloser         io.Closer
	mux               *http.ServeMux
	HttpListenAddress string
}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	fmt.Println(req)
	r.mux.ServeHTTP(res, req)
}

func (r *Router) Close() error {
	if r.Router != nil {
		r.Router.Close()
	}
	if r.wsCloser != nil {
		if err := r.wsCloser.Close(); err != nil {
			return err
		}
	}
	if r.tcpCloser != nil {
		if err := r.tcpCloser.Close(); err != nil {
			return err
		}
	}
	return nil
}

func NewRouter(realm string, wsHost string, wsPort int, tcpHost string, tcpPort int) (*Router, error) {
	routerConfig := &router.Config{
		RealmConfigs: []*router.RealmConfig{
			{
				URI:           wamp.URI(realm),
				AnonymousAuth: true,
				AllowDisclose: true,
			},
		},
	}
	nxr, err := router.NewRouter(routerConfig, nil)
	if err != nil {
		return nil, err
	}

	newRouter := &Router{
		Router: nxr,
		Realm:  realm,
		mux:    http.NewServeMux(),
	}

	// Create servers
	wss := router.NewWebsocketServer(nxr)
	wss.Upgrader.EnableCompression = true
	wss.Upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	wss.EnableTrackingCookie = true
	wss.KeepAlive = 30 * time.Second
	var rss *router.RawSocketServer
	if tcpPort > 0 {
		rss = router.NewRawSocketServer(nxr)
	}
	// Start servers
	wsAddr := fmt.Sprintf("%s:%d", wsHost, wsPort)
	newRouter.HttpListenAddress = wsAddr
	newRouter.mux.Handle("/", wss)
	//wsCloser, err := wss.ListenAndServe(wsAddr)
	//if err != nil {
	//	return nil, err
	//}
	//newRouter.wsCloser = wsCloser
	log.Printf("Websocket server listening on ws://%s/", wsAddr)

	if tcpPort > 0 {
		tcpAddr := fmt.Sprintf("%s:%d", tcpHost, tcpPort)
		tcpCloser, err := rss.ListenAndServe("tcp", tcpAddr)
		if err != nil {
			return nil, err
		}
		newRouter.tcpCloser = tcpCloser
		log.Printf("RawSocket TCP server listening on tcp://%s/", tcpAddr)
	}

	return newRouter, nil
}

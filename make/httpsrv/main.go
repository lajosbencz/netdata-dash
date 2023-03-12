package main

import (
	"crypto/tls"
	"embed"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"context"
)

//go:embed web
var webEmbed embed.FS

func serveStatic(fs http.FileSystem, path string, w http.ResponseWriter, r *http.Request) bool {
	f, err := fs.Open(path)
	if err == nil {
		defer f.Close()
		fi, err := f.Stat()
		if err == nil {
			if !fi.IsDir() {
				http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
				return true
			}
		}
	}
	return false
}

const (
	addrHttp  = "localhost:1337"
	addrHttps = "localhost:1338"
)

func main() {
	publicPath := "./make/httpsrv/public"
	publicHandler := http.Dir(publicPath)
	webHandler := http.FS(webEmbed)
	mux := http.NewServeMux()
	mux.HandleFunc("/ws/", wsHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path += "index.html"
		}
		if !serveStatic(publicHandler, path, w, r) {
			if !serveStatic(webHandler, "/web"+path, w, r) {
				http.NotFound(w, r)
			}
		}
	})

	httpServer := &http.Server{Addr: addrHttp, Handler: mux}

	tlsCfg := &tls.Config{}
	tlsCfg.NextProtos = []string{"http/1.1"}
	tlsCfg.Certificates = make([]tls.Certificate, 1)
	cert, err := GenX509KeyPair("localhost", "hu", "dev", "dev", 60)
	if err != nil {
		log.Fatalln(err)
	}
	tlsCfg.Certificates[0] = cert
	ln, err := net.Listen("tcp", addrHttps)
	if err != nil {
		log.Fatalln(err)
	}
	httpsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener), 60 * time.Minute}, tlsCfg)
	httpsServer := &http.Server{Addr: addrHttps, Handler: mux}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go httpServer.ListenAndServe()
	go httpsServer.Serve(httpsListener)

	log.Printf("listening on http://%s\n", addrHttp)
	log.Printf("listening on https://%s\n", addrHttps)

	<-shutdown

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		httpServer.Shutdown(ctx)
	}()
	go func() {
		defer wg.Done()
		httpsServer.Shutdown(ctx)
	}()
	wg.Wait()
	cancel()
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ws endpoint goes here"))
}

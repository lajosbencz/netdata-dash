package main

import (
	"embed"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/flosch/pongo2/v6"
)

//go:embed public
var publicEmbed embed.FS

const (
	staticPath = "public"
)

func newHttp() func(w http.ResponseWriter, r *http.Request) {
	publicPath := "./" + staticPath
	publicRutimeHandler := http.Dir(publicPath)
	publicEmbedHandler := http.FS(publicEmbed)
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path += "index.html"
		}
		if !serveStaticFile(publicRutimeHandler, path, w, r) {
			if !serveStaticFile(publicEmbedHandler, "/"+staticPath+path, w, r) {
				http.NotFound(w, r)
			}
		}
	}
}

func serveStaticFile(fs http.FileSystem, path string, w http.ResponseWriter, r *http.Request) bool {
	f, err := fs.Open(path)
	if err == nil {
		defer f.Close()
		fi, err := f.Stat()
		if err == nil {
			if !fi.IsDir() {
				if strings.HasSuffix(fi.Name(), ".html") {
					content, err := io.ReadAll(f)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return true
					}
					tpl := pongo2.Must(pongo2.FromBytes(content))
					if err := tpl.ExecuteWriter(pongo2.Context{}, w); err != nil {
						log.Println(err)
					}
					return true
				} else {
					http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
				}
			}
		}
	}
	return false
}

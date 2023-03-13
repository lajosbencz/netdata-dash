package main

import (
	"embed"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/flosch/pongo2/v6"
	loader "github.com/nathan-osman/pongo2-embed-loader"
)

//go:embed template/*
var templateEmbed embed.FS

//go:embed public/*
var publicEmbed embed.FS

const (
	publicDir = "public"
	tvDir     = "tv"
)

type myHttp struct {
	embedHandler    http.FileSystem
	runtimeHandler  http.FileSystem
	tplIndex        *pongo2.Template
	tplIndexContent *pongo2.Template
}

func newHttp() (*myHttp, error) {
	publicPath := "./" + publicDir
	publicRuntimeHandler := http.Dir(publicPath)
	publicEmbedHandler := http.FS(publicEmbed)
	templateSet := pongo2.NewSet("", &loader.Loader{Content: templateEmbed})
	tplIndex, err := templateSet.FromFile("template/index.html")
	if err != nil {
		return nil, err
	}
	tplContent, err := templateSet.FromFile("template/index_content.html")
	if err != nil {
		return nil, err
	}
	return &myHttp{
		embedHandler:    publicEmbedHandler,
		runtimeHandler:  publicRuntimeHandler,
		tplIndex:        tplIndex,
		tplIndexContent: tplContent,
	}, nil
}

func (t *myHttp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" || path == "/index.html" {
		if err := t.tplIndex.ExecuteWriter(pongo2.Context{}, w); err != nil {
			// http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
		}
		return
	}
	if strings.HasPrefix(path, "/"+tvDir+"/") {
		tvPath := path[3:]
		if tvPath == "/" {
			tvPath += "index.html"
		}
		content, err := os.ReadFile("./" + tvDir + tvPath)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}
			content = []byte("<pre>" + err.Error() + "</pre>")
		}
		if err := t.tplIndexContent.ExecuteWriter(pongo2.Context{
			"content": string(content),
		}, w); err != nil {
			log.Println(err)
		}
		return
	}
	if !serveStaticFile(t.runtimeHandler, path, w, r) {
		if !serveStaticFile(t.embedHandler, "/"+publicDir+path, w, r) {
			http.NotFound(w, r)
			return
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
				} else {
					http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
				}
				return true
			}
		}
	}
	return false
}

// +build ignore

package main

import (
	"fmt"
	"net/http"

	"github.com/MJKWoolnough/downloader"

	"github.com/MJKWoolnough/downloader/cache"
	_ "github.com/MJKWoolnough/downloader/sites/youtube"
)

var fileCache *cache.Cache

func main() {
	fileCache = cache.NewCache("/home/michael/temp/dlcache/")
	http.ListenAndServe(":8080", http.HandlerFunc(proxy))
}

func proxy(w http.ResponseWriter, r *http.Request) {
	url := r.RequestURI
	if url[0] == '/' {
		url = url[1:]
	}
	req, err := downloader.DoRequest(url)
	if err != nil {
		if _, ok := err.(downloader.NoRequest); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		fmt.Println(err)
		return
	}
	d := req.Downloaders[0]

	c, err := fileCache.Get(d.UID, d.Sources[0])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Add("Content-Disposition", fmt.Sprintf("inline; filename=%q", req.Filename))
	http.ServeContent(w, r, req.Filename, d.LastModified, c)
}

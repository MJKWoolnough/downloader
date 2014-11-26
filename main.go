// +build ignore

package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MJKWoolnough/downloader"

	_ "github.com/MJKWoolnough/downloader/sites/youtube"
)

func main() {
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
	w.Header().Add("Content-Disposition", fmt.Sprintf("inline; filename=%q", req.Title+".mp4"))
	d := req.Downloaders[0]
	//c := cache.Get(d.UID, d.Sources, offset, length)
	//io.Copy(w, c)
	//c.Close()
	w.Header().Add("Content-Length", fmt.Sprintf("%d", d.Size))
	w.Header().Add("Last-Modified", d.LastModified.Format(time.RFC850))
	w.Header().Add("Content-Type", d.MimeType)
	//read := 0
	//errs := 0
	io.Copy(w, d.Sources[0])
	/*	for read < d.Size && errs < len(d.Sources) {
		n, err := io.Copy(w, d.Sources[errs])
		d.Sources[errs].Close()
		read += int(n)
		if err != nil {
			errs++
			fmt.Println(err)
		}
	}*/
}

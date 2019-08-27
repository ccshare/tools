package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func serveProxy(url *url.URL, w http.ResponseWriter, r *http.Request) {

	proxy := httputil.NewSingleHostReverseProxy(url)

	r.URL.Host = url.Host
	r.URL.Scheme = url.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = url.Host

	proxy.ServeHTTP(w, r)
}

func main() {
	server := flag.String("s", "http://172.16.3.50:9020", "upstream server")
	port := flag.Int("p", 80, "listen port")
	flag.Parse()

	url, err := url.Parse(*server)
	if err != nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveProxy(url, w, r)
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Println(err)
	}
}

package main

// test
import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var client = http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 5 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   1000,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   8 * time.Second,
		ExpectContinueTimeout: 8 * time.Second,
	},
	Timeout: 10 * time.Second,
}

func serveProxy(url *url.URL, w http.ResponseWriter, r *http.Request) {
	fmt.Println("serve by proxy")
	//r.Host = url.Host
	r.URL.Host = url.Host
	r.URL.Scheme = url.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ModifyResponse = func(resp *http.Response) error {
		fmt.Println("status: ", resp.Status)
		fmt.Println(resp.ContentLength)
		for h, v := range resp.Header {
			fmt.Println(h, ": ", v)
		}
		fmt.Println("proxy --->")
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		amzID := resp.Header.Get("X-Amz-Id-2")
		if amzID != "" {
			resp.Header.Del("x-amz-id-2")
			resp.Header["x-amz-id-2"] = []string{amzID}
		}
		resp.Header["x-proxy-id"] = []string{"x-proxy-id-lower"}
		resp.Body = ioutil.NopCloser(bytes.NewReader(body))
		return nil
	}
	proxy.ServeHTTP(w, r)
}

func serveRequest(url *url.URL, w http.ResponseWriter, r *http.Request) {
	fmt.Println("serve by request")
	req := r.Clone(context.Background())
	req.URL.Host = url.Host
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("client Do error: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	for h, v := range resp.Header {
		fmt.Println(h, ": ", v)
	}
	fmt.Println("request --->")

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		fmt.Println("io copy error: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

}

func main() {
	server := flag.String("s", "http://172.16.3.98:9020", "upstream server")
	req := flag.Bool("req", false, "use buildin reverseproxy")
	addr := flag.String("addr", ":9033", "serve address")
	flag.Parse()

	url, err := url.Parse(*server)
	if err != nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if *req {
			serveProxy(url, w, r)
		} else {
			serveRequest(url, w, r)
		}
	})

	if err := http.ListenAndServe(*addr, nil); err != nil {
		fmt.Println(err)
	}
}

package main

// test
import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/dgraph-io/badger/v2"
)

var (
	db *badger.DB
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

// defaultTransport Transport for gateway
var defaultTransport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          128,
	MaxIdleConnsPerHost:   128,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

func serveProxy(realURL *url.URL, w http.ResponseWriter, r *http.Request) {
	realURL.Path = r.URL.Path
	proxy := httputil.ReverseProxy{
		Transport: defaultTransport,
		Director: func(req *http.Request) {
			req.URL = realURL
			req.Host = realURL.Host
			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
		},
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode == http.StatusOK && resp.Request.Method == http.MethodGet {
			//visits, shouldCache := cacheIns.ShouldCache(r.Method, bucket.ID, cacheKey, resp.Header.Get("Content-Length"))
			shouldCache := false
			if shouldCache {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Printf("read upstream body error %s", err)
					return err
				}
				resp.Body = ioutil.NopCloser(bytes.NewReader(body))
				// write cache
			} else {
				log.Println("success and not cache")
			}
		}
		return nil
	}
	proxy.ServeHTTP(w, r)
}

func main() {
	server := flag.String("s", "http://192.168.55.2:9000", "upstream server")
	ddir := flag.String("d", "/tmp/db", "db dir")
	addr := flag.String("addr", ":80", "serve address")
	flag.Parse()

	url, err := url.Parse(*server)
	if err != nil {
		fmt.Println(err)
		return
	}

	opts := badger.DefaultOptions(*ddir)
	db, err = badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveProxy(url, w, r)
	})

	if err := http.ListenAndServe(*addr, nil); err != nil {
		fmt.Println(err)
	}
}

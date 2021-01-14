package main

// test
import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"
)

var (
	db          *badger.DB
	redisClient *redis.Client
	redisOnce   sync.Once

	redisAddr string
)

// ParseRedisAddr parse redis passwd,ip,port,db from addr
func ParseRedisAddr(addr string) (host, passwd string, db int, err error) {
	var u *url.URL
	u, err = url.ParseRequestURI(addr)
	if err != nil {
		return
	}

	if u.User != nil {
		var exists bool
		passwd, exists = u.User.Password()
		if !exists {
			passwd = u.User.Username()
		}
	}

	host = u.Host
	parts := strings.Split(u.Path, "/")
	if len(parts) == 1 {
		db = 0 //default redis db
	} else {
		db, err = strconv.Atoi(parts[1])
		if err != nil {
			db, err = 0, nil //ignore err here
		}
	}

	return
}

// RedisGetInstance init and return a redis client
func RedisGetInstance() *redis.Client {
	if redisClient != nil {
		return redisClient
	}
	redisOnce.Do(func() {
		host, passwd, db, err := ParseRedisAddr(redisAddr)
		if err != nil {
			panic(fmt.Sprintf("parse redis addr %s, error %s", redisAddr, err))
		}
		redisClient = redis.NewClient(&redis.Options{
			Addr:     host,
			Password: passwd,
			DB:       db,
		})
	})
	return redisClient
}

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

func serveDBProxy(realURL *url.URL, w http.ResponseWriter, r *http.Request) {
	read := db.NewTransaction(false)
	item, err := read.Get([]byte(r.URL.Path))
	if err == nil {
		if v, e := item.ValueCopy(nil); e == nil {
			log.Println("got from cache")
			w.Write(v)
			return
		}
	}

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
			shouldCache := true
			if shouldCache {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Printf("read upstream body error %s", err)
					return err
				}
				resp.Body = ioutil.NopCloser(bytes.NewReader(body))
				// write cache
				txn := db.NewTransaction(true) // Read-write txn
				err = txn.SetEntry(badger.NewEntry([]byte(r.URL.Path), body).WithTTL(1 * time.Minute))
				if err != nil {
					panic(err)
				}
				err = txn.Commit()
				if err != nil {
					panic(err)
				}
				log.Println("success and cached")
			} else {
				log.Println("success and not cache")
			}
		}
		return nil
	}
	proxy.ServeHTTP(w, r)
}

func shouldCache(method, cacheKey, contentLen string) (int, bool) {
	return 0, true
}

func serveRedisProxy(realURL *url.URL, w http.ResponseWriter, r *http.Request) {
	cacheKey := r.URL.Path
	cacheIns := RedisGetInstance()
	if tmpBuffer, err := cacheIns.Get(cacheKey).Bytes(); err == nil {
		log.Printf("hit cache of %s", cacheKey)
		w.Header().Set("X-Redis-Cache", "1")
		w.Write(tmpBuffer)
		return
	}

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
			counter, shouldCache := shouldCache(r.Method, cacheKey, resp.Header.Get("Content-Length"))
			if shouldCache {

				for h, v := range resp.Header {
					w.Header().Set(h, v[0])
				}

				data := make([]byte, 4096)
				n, err := resp.Body.Read(data)
				if err != nil && err != io.EOF {
					log.Printf("proxy read resp body failed %s", err)
					return err
				}
				log.Printf("proxy read resp body %v, %v", n, err)
				w.Write(data[0:n])
				if ret := cacheIns.Set(cacheKey, data[0:n], 5*time.Second).Err(); ret == nil {
					log.Printf("proxy success and cached")
				} else {
					log.Printf("proxy success and cache failed")
				}
			} else {
				log.Printf("proxy success and not cache, %v ", counter)
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
	flag.StringVar(&redisAddr, "redis", "redis://192.168.55.2:6379/3", "redis service address")
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

	router := mux.NewRouter()
	router.HandleFunc("/dynamic/{key:.*}", func(w http.ResponseWriter, r *http.Request) {
		serveDBProxy(url, w, r)
	})

	router.HandleFunc("/open/{key:.*}", func(w http.ResponseWriter, r *http.Request) {
		serveRedisProxy(url, w, r)
	})
	http.Handle("/", router)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		fmt.Println(err)
	}
}

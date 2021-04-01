package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var httpClient = http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   1000,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   30 * time.Second,
		ExpectContinueTimeout: 30 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		DisableKeepAlives:     true,
	},
	Timeout: 30 * time.Second,
}

func main() {
	addr := flag.String("addr", "http://127.0.0.1/open/hosts", "server")
	flag.Parse()
	fmt.Println(*addr)

	req, err := http.NewRequest(http.MethodGet, *addr, nil)
	if err != nil {
		fmt.Println(err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	err = resp.Body.Close()
	fmt.Println("close: ", err)

	d, _ := ioutil.ReadAll(resp.Body)

	fmt.Printf("%s", d)
}

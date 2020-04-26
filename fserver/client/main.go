package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	addr := flag.String("addr", "http://127.0.0.1", "server url")
	flag.Parse()
	fmt.Println(*addr)

	req, err := http.NewRequest(http.MethodGet, *addr, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("x-htest", " ")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	d, _ := ioutil.ReadAll(resp.Body)

	fmt.Printf("%s", d)
}

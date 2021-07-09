package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	interval    int
	nBytes      int64
	idleTimeout = 30
	addr        string
	output      string
)

func main() {
	flag.StringVar(&addr, "addr", "", "URL go get")
	flag.StringVar(&output, "o", "a.out", "output file")
	flag.Int64Var(&nBytes, "n", 64, "bytes to download one time")
	flag.IntVar(&idleTimeout, "idle-timeout", 60, "http idle timeout after download in seconds")
	flag.IntVar(&interval, "i", 1, "download interval time in seconds")
	flag.Parse()

	httpClient := http.Client{}

	resp, err := httpClient.Get(addr)
	if err != nil {
		fmt.Println("get error: ", err)
		return
	}
	defer resp.Body.Close()
	f, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println("open file error: ", err)
		return
	}
	defer f.Close()

	for {
		n, err := io.CopyN(f, resp.Body, nBytes)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("download error: %s", err)
			return
		}
		time.Sleep(time.Duration(interval) * time.Second)
		log.Printf("download %v bytes", n)
	}

	log.Printf("download finish, enter idle mode")
	time.Sleep(time.Duration(idleTimeout) * time.Second)
	log.Printf("done")

}

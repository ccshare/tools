package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	interval    int
	nBytes      int64
	nFiles      int
	idleTimeout = 30
	addr        string
	output      string
)
var httpClient = http.Client{}

func downloadFile(addr, filename string) error {
	resp, err := httpClient.Get(addr)
	if err != nil {
		return fmt.Errorf("get error: %w", err)
	}
	for k, v := range resp.Header {
		fmt.Printf("%s: %v\n", k, v)
	}
	defer resp.Body.Close()
	f, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("open local file error: %w", err)
	}
	defer f.Close()

	for {
		n, err := io.CopyN(f, resp.Body, nBytes)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("write local file error: %w", err)
		}
		time.Sleep(time.Duration(interval) * time.Second)
		log.Printf("download %v bytes", n)
	}
	return nil
}

func main() {
	flag.StringVar(&addr, "addr", "http://127.0.0.1/open/f", "URLs go get")
	flag.StringVar(&output, "o", "a.out", "output files")
	flag.IntVar(&nFiles, "fn", 2, "downlaod file num")
	flag.Int64Var(&nBytes, "n", 128, "bytes to download one time")
	flag.IntVar(&idleTimeout, "idle-timeout", 60, "http idle timeout after download in seconds")
	flag.IntVar(&interval, "i", 1, "download interval time in seconds")
	flag.Parse()

	for i := 1; i <= nFiles; i++ {
		iStr := strconv.Itoa(i)
		if err := downloadFile(addr+iStr, output+iStr); err != nil {
			log.Println("download error: ", err)
			return
		}
		log.Printf("%v download finished", i)
		time.Sleep(time.Duration(interval) * time.Second)
	}

	log.Printf("all finished, enter idle mode")
	time.Sleep(time.Duration(idleTimeout) * time.Second)
	log.Printf("all done")

}

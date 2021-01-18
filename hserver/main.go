package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/henvic/httpretty"
)

var (
	port uint
	root string
)

func main() {
	flag.UintVar(&port, "p", 8080, "listen port")
	flag.StringVar(&root, "r", ".", "web root dir")
	flag.Parse()

	logger := &httpretty.Logger{
		Time:           true,
		TLS:            true,
		RequestHeader:  true,
		RequestBody:    true,
		ResponseHeader: true,
		ResponseBody:   true,
		Colors:         true, // erase line if you don't like colors
	}

	addr := fmt.Sprintf(":%v", port)
	fmt.Printf("Listen %s\n", addr)

	if err := http.ListenAndServe(addr, logger.Middleware(http.FileServer(http.Dir(root)))); err != http.ErrServerClosed {
		fmt.Fprintln(os.Stderr, err)
	}
}

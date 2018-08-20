package main

import (
	"fmt"
	"flag"
	"net/http"
)


func main() {
	host := flag.String("host", "0.0.0.0", "Listen address")
	port := flag.Int("port", 80, "Listen port")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
                fmt.Println(*r)
        })

	flag.Parse()

	addr := fmt.Sprintf("%s:%d", *host, *port)
	fmt.Printf("Listen on %s\n", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println(err)
	}
}

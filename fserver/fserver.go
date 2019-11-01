package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func fserver(host string, port int) {
	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Listen on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, http.FileServer(http.Dir("."))))
}

func server(host string, port int, content string) {
	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Listen on %s\n", addr)
	body := []byte(content)

	http.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) {
		w.Write(body)
	})

	http.HandleFunc("/a/b", func(w http.ResponseWriter, r *http.Request) {
		w.Write(body)
	})

	log.Fatal(http.ListenAndServe(addr, nil))
}

func router(host string, port int, content string) {
	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Listen on %s\n", addr)
	body := []byte(content)
	router := httprouter.New()

	router.Handle("GET", "/a", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Write(body)
	})

	router.Handle("GET", "/a/:id", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Write(body)
	})

	log.Fatal(http.ListenAndServe(addr, router))
}

func main() {
	host := flag.String("host", "", "Listen address")
	port := flag.Int("port", 80, "Listen port")
	data := flag.String("data", "default http body", "http response body")

	flag.Parse()

	//fserver(*host, *port)

	//server(*host, *port, *data)

	router(*host, *port, *data)
}

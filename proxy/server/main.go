package main

// test
import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":80", "serve addr")

	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-request-id", "x-request-id-canonical")
		w.Header()["x-amz-id-2"] = []string{"x-amz-id-2-lower"}
		w.Write([]byte("ok"))
	})

	if err := http.ListenAndServe(*addr, nil); err != nil {
		fmt.Println(err)
	}
}

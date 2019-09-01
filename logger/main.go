package main

import (
	"flag"
	"fmt"
	"net/http"

	log1 "github.com/sirupsen/logrus"
	zap "go.uber.org/zap"
)

var content = "abcdefghijklmnopqrstuvwxyz123456"

func main() {
	port := flag.Int("p", 80, "listen port")
	flag.Parse()

	log1.SetFormatter(&log1.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	log2, err := zap.NewProduction()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer log2.Sync()

	http.HandleFunc("/log1", func(w http.ResponseWriter, r *http.Request) {
		i := 0
		log1.Info("log1 info index: ", i)
		i++
		log1.Info("log1 info index: ", i)
		i++
		log1.Info("log1 info index: ", i)
		i++
		log1.Info("log1 info index: ", i)
		w.Write([]byte(content))
	})

	http.HandleFunc("/log2", func(w http.ResponseWriter, r *http.Request) {
		i := 0
		log2.Info("log2 info ",
			zap.Int("index", i),
		)
		i++
		log2.Info("log2 info ",
			zap.Int("index", i),
		)
		i++
		log2.Info("log2 info ",
			zap.Int("index", i),
		)
		i++
		log2.Info("log2 info ",
			zap.Int("index", i),
		)
		w.Write([]byte(content))
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Println(err)
	}
}

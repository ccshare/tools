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

	// logrus
	//log1.SetReportCaller(true)
	//log1.SetFormatter(&log1.JSONFormatter{})
	log1.SetFormatter(&log1.TextFormatter{DisableColors: true})

	// zap
	log2, err := zap.NewProduction()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer log2.Sync()

	http.HandleFunc("/log1", func(w http.ResponseWriter, r *http.Request) {
		i := 0
		log1.WithFields(log1.Fields{
			"name":  "walrus",
			"index": i,
		}).Info("log1 sample info")

		i++
		log1.WithFields(log1.Fields{
			"name":  "walrus",
			"index": i,
		}).Info("log1 sample info")

		i++
		log1.WithFields(log1.Fields{
			"name":  "walrus",
			"index": i,
		}).Info("log1 sample info")

		i++
		log1.WithFields(log1.Fields{
			"name":  "walrus",
			"index": i,
		}).Info("log1 sample info")

		w.Write([]byte(content))
	})

	http.HandleFunc("/log2", func(w http.ResponseWriter, r *http.Request) {
		i := 0
		log2.Info("log2 sample info",
			zap.Int("index", i),
			zap.String("name", "walrus"),
		)
		i++
		log2.Info("log2 sample info",
			zap.Int("index", i),
			zap.String("name", "walrus"),
		)
		i++
		log2.Info("log2 sample info",
			zap.Int("index", i),
			zap.String("name", "walrus"),
		)
		i++
		log2.Info("log2 sample info",
			zap.Int("index", i),
			zap.String("name", "walrus"),
		)
		w.Write([]byte(content))
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Println(err)
	}
}

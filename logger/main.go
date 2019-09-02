package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/sirupsen/logrus"
	zap "go.uber.org/zap"
)

var content = "abcdefghijklmnopqrstuvwxyz123456"

func main() {
	port := flag.Int("p", 80, "listen port")
	num := flag.Int("n", 1, "log number per request")
	flag.Parse()

	http.HandleFunc("/loglog", func(w http.ResponseWriter, r *http.Request) {
		for i := 0; i < *num; i++ {
			log.Println("loglog sample info, index=", i)
		}
		w.Write([]byte(content))
	})

	// logrus
	//logrus.SetReportCaller(true)
	//logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetFormatter(&logrus.TextFormatter{DisableColors: true})
	http.HandleFunc("/logrus", func(w http.ResponseWriter, r *http.Request) {
		for i := 0; i < *num; i++ {
			logrus.WithFields(logrus.Fields{
				"name":  "walrus",
				"index": i,
			}).Info("logrus sample info")
		}
		w.Write([]byte(content))
	})

	// zap
	logzap, err := zap.NewProduction()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer logzap.Sync()
	http.HandleFunc("/logzap", func(w http.ResponseWriter, r *http.Request) {
		for i := 0; i < *num; i++ {
			logzap.Info("logzap sample info",
				zap.Int("index", i),
				zap.String("name", "walrus"),
			)
		}
		w.Write([]byte(content))
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Println(err)
	}
}

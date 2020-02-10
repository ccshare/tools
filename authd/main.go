package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/globalsign/mgo"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger
var version = "unknown"

func loginHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func pingHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

// init logger
func initLogger(debug bool) (logger *zap.Logger) {
	var err error
	var zcfg zap.Config
	if debug {
		zcfg = zap.NewDevelopmentConfig()
	} else {
		zcfg = zap.NewProductionConfig()
		// Change default(1578990857.105345) timeFormat to 2020-01-14T16:35:34.851+0800
		zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}
	if os.Getenv("LOGGER") == "file" {
		filename := filepath.Base(os.Args[0])
		zcfg.OutputPaths = []string{
			filepath.Join("/tmp", filename),
		}
	}

	logger, err = zcfg.Build()
	if err != nil {
		panic(fmt.Sprintf("initLooger error %s", err))
	}

	zap.ReplaceGlobals(logger)
	return
}

func initDB(addr, dbName, table string) {
	dI, err := mgo.ParseURL(addr)
	if err != nil {
		panic(err)
	}

	session, err := mgo.DialWithInfo(dI)
	if err != nil {
		panic(err)
	}

	db := session.DB(dbName)
	collection := db.C(table)

	collection.Find("")
}

func main() {
	port := flag.Uint("port", 80, "port")
	debug := flag.Bool("debug", false, "debug")
	ver := flag.Bool("version", false, "version")
	flag.Parse()

	if *ver {
		fmt.Println(version)
		return
	}

	logger = initLogger(*debug)
	defer logger.Sync()

	router := httprouter.New()
	router.POST("/login", loginHandler)
	router.GET("/ping/:id", pingHandler)

	addr := fmt.Sprintf(":%d", *port)

	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Fatal("ListenAndServe",
			zap.String("addr", addr),
			zap.String("err", err.Error()),
		)
	}
}

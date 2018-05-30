package main

import (
	"encoding/json"
	"flag"
	"github.com/golang/glog"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type tokenStruct struct {
	Token string `json:"token"`
}

type pblockResult struct {
	// {"result":"Upload completed"}
	Result string `json:"result"`
}

var fileServer *FileServer

func handleToken(w http.ResponseWriter, r *http.Request) {
	// http://host:port/token?entryKey=$key&entryOp=put
	// http://host:port/token?entryKey=$key&entryOp=get
	token := tokenStruct{}
	defer r.Body.Close()
	defer func() {
		jsdata, err := json.Marshal(token)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Write(jsdata)
		}
	}()
	if r.Method == "GET" || r.Method == "POST" {
		r.ParseForm()
		key := r.Form["entryKey"]
		op := r.Form["entryOp"]

		if len(key) == 1 && len(op) == 1 && (op[0] == "get" || op[0] == "put") {
			respStr, err := fileServer.Token(key[0], op[0])
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				token.Token = "gen token failed"
			} else {
				token.Token = respStr
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			token.Token = "unknow parameter"
		}
	} else {
		w.WriteHeader(http.StatusNotImplemented)
		token.Token = "Method not support"
	}
}

func handlePblocks(w http.ResponseWriter, r *http.Request) {
	// http://host:port/pblocks/$key?token=$ptoken
	result := pblockResult{}
	defer r.Body.Close()
	if r.Method == "GET" || r.Method == "POST" {
		r.ParseForm()
		keys := strings.Split(r.URL.Path, "/")
		token := r.Form["token"]
		if len(token) == 0 || len(keys) != 3 {
			result.Result = "handle Get File, error parameters"
		} else {
			if r.Method == "GET" {
				data, err := fileServer.Download(token[0], keys[2])
				if err != nil {
					w.WriteHeader(404)
					result.Result = "handle Get File failed"
				} else {
					w.Header().Set("content-type", "application/octet-stream")
					w.Write(data)
					return
				}
			} else {
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					log.Println("Read req.Body", err)
					result.Result = "read req body error"
				} else {
					data, err := fileServer.Upload(token[0], keys[2], body)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						result.Result = "Upload error"
						fmt.Println("Upload error: ", err)
					} else {
						result.Result = data
					}
				}
			}
		}
	} else {
		result.Result = "method not support"
	}
	jsdata, err := json.Marshal(result)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Write(jsdata)
	}
}

func server(addr string, port int) error {
	addrPort := fmt.Sprintf(":%d", port)
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/token", handleToken)
	http.HandleFunc("/pblocks/", handlePblocks)
	err := http.ListenAndServe(addrPort, nil)
	if err != nil {
		log.Println(err)
	}

	return nil
}

func main() {
	const VERSION = "version: 1.0.1"
	address := flag.String("address", "0.0.0.0", "listen ip address")
	port := flag.Int("port", 3000, "port")
	db := flag.String("db", "cwd", "db path")
	debug := flag.Bool("debug", false, "enable/disable debug mode")
	ignore := flag.Bool("i", false, "ignore failed validation")
	version := flag.Bool("version", false, "show version")

	flag.Parse()
	if *version == true {
		fmt.Printf("%s  %s\n", os.Args[0], VERSION)
		return
	}

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	logdir := filepath.Join(pwd, "log")
	var dbpath string
	if *db == "cwd" {
		dbpath = filepath.Join(pwd, "filestore")
	} else {
		dbpath = filepath.Join(*db, "filestore")
	}
	_ = os.MkdirAll(logdir, 0755)
	_ = os.MkdirAll(dbpath, 0755)

	rand.Seed(time.Now().Unix())
	randValue := rand.Intn(1)

	logFilename := fmt.Sprintf("%s-%05d.log", os.Args[0], randValue)
	logFilename = filepath.Join(logdir, logFilename)
	logFile, logErr := os.OpenFile(logFilename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if logErr != nil {
		fmt.Println("Fail to OpenFile", logErr)
		os.Exit(1)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	glog.Info("Start fserver [test glog info]")
	glog.Warning("Start fserver [test glog warning]")
	glog.Error("Start fserver [test glog error]")
	glog.V(1).Infoln("level 1")
	glog.V(2).Infoln("level 2")
	defer glog.Flush()

	serverURL := fmt.Sprintf("http://%s:%d", *address, *port)

	fmt.Printf("server:%s, logfile:%s\n", serverURL, logFilename)
	log.Printf("server:%s, logfile:%s debug: %v, db:%s, ignore:%v\n", serverURL, logFilename, *debug, dbpath, *ignore)

	fileServer = NewFileServer(dbpath)

	server(*address, *port)

}

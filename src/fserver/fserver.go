package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type tokenStruct struct {
	Token string
	Error bool
}

type pblockResult struct {
	// {"result":"Upload completed"}
	Result string
}

func handleToken(w http.ResponseWriter, r *http.Request) {
	// http://host:port/token?entryKey=$key&entryOp=put
	// http://host:port/token?entryKey=$key&entryOp=get
	token := tokenStruct{Error: true}
	if r.Method == "GET" || r.Method == "POST" {
		r.ParseForm()
		key := r.Form["entryKey"]
		op := r.Form["entryOp"]

		if len(key) == 0 || len(op) == 0 {
			token.Token = "error parameters"
		} else if op[0] == "get" || op[0] == "put" {
			respStr, err := Token(key[0], op[0])
			if err != nil {
				token.Token = "gen token failed"
			} else {
				token.Error = false
				token.Token = respStr
			}
		} else {
			token.Token = "unknow op"
		}
	} else {
		token.Token = "Method not support"
	}
	jsdata, err := json.Marshal(token)
	if err != nil {
		log.Println(err)
	} else {
		w.Write(jsdata)
	}
}

func handlePblocks(w http.ResponseWriter, r *http.Request) {
	// http://host:port/pblocks/$key?token=$ptoken
	result := pblockResult{}
	if r.Method == "GET" || r.Method == "POST" {
		r.ParseForm()
		keys := strings.Split(r.URL.Path, "/")
		token := r.Form["token"]
		if len(token) == 0 || len(keys) != 3 {
			result.Result = "handle Get File, error parameters"
		} else {
			if r.Method == "GET" {
				Download(keys[2])
				w.Write([]byte("handle Get File"))
				return
			}
			data, err := Upload(keys[2])
			if err != nil {
				result.Result = "Upload error"
			} else {
				result.Result = data
			}
		}
	} else {
		w.Write([]byte("not support"))
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
	port := flag.Int("port", 3003, "port")
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
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	serverURL := fmt.Sprintf("http://%s:%d", *address, *port)

	fmt.Printf("server:%s, logfile:%s\n", serverURL, logFilename)
	log.Printf("server:%s, logfile:%s debug: %v, db:%s, ignore:%v\n", serverURL, logFilename, *debug, dbpath, *ignore)

	server(*address, *port)

}

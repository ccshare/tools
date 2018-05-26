package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type tokenStruct struct {
	Token string
}

func upload(serverURL string, entryKey string, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()
	// http://${host_port}/token?entryKey=$key&entryOp=put
	tokenURL := fmt.Sprintf("%s/token?entryKey=%s&entryOp=put", serverURL, entryKey)
	log.Println(tokenURL)
	tokenResp, err := http.Post(tokenURL, "application/json", strings.NewReader(""))
	if err != nil {
		log.Fatal(err)
	}

	defer tokenResp.Body.Close()
	tokenBody, err := ioutil.ReadAll(tokenResp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(tokenBody))
	token := tokenStruct{}
	err = json.Unmarshal(tokenBody, &token)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	log.Printf("put token: %s", token.Token)

	// http://${host_port}/pblocks/$key?token=$ptoken Content-Type: application/octet-stream
	postURL := fmt.Sprintf("%s/pblocks/%s?token=%s", serverURL, entryKey, token.Token)
	postResp, err := http.Post(postURL, "application/octet-stream", file)
	if err != nil {
		log.Fatal(err)
	}

	defer postResp.Body.Close()
	postBody, err := ioutil.ReadAll(postResp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(postBody))
}

func download(serverURL string, entryKey string, filename string) {
	// http://${host_port}/token?entryKey=$key&entryOp=get
	tokenURL := fmt.Sprintf("%s/token?entryKey=%s&entryOp=get", serverURL, entryKey)
	log.Println(tokenURL)
	tokenResp, err := http.Post(tokenURL, "application/json", strings.NewReader(""))
	if err != nil {
		log.Fatal(err)
	}

	defer tokenResp.Body.Close()
	tokenBody, err := ioutil.ReadAll(tokenResp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(tokenBody))
	token := tokenStruct{}
	err = json.Unmarshal(tokenBody, &token)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	log.Printf("get token: %s", token.Token)

}

func main() {
	fmt.Println("start")
	host := flag.String("host", "localhost", "hostname")
	port := flag.Int("port", 3000, "port")
	key := flag.String("key", "key", "key")
	ufile := flag.String("ufile", "/tmp/ufile", "upload filename")
	sfile := flag.String("save", "/tmp/downfile", "save download filename")
	num := flag.Int("num", 1024, "upload/download times")
	debug := flag.Bool("debug", false, "enable/disable debug mode")
	ignore := flag.Bool("i", false, "ignore failed validation")

	flag.Parse()
	rand.Seed(time.Now().Unix())
	randKey := fmt.Sprintf("%s-%04d", *key, rand.Intn(4096))

	serverURL := fmt.Sprintf("http://%s:%d", *host, *port)
	fmt.Printf("key:%s, ufile:%s, dfile:%s debug: %v, num:%d, ignore:%v\n", *key, *ufile, *sfile, *debug, *num, *ignore)

	upload(serverURL, randKey, *ufile)
	download(serverURL, randKey, *sfile)

}

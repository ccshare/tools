package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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

func token(serverURL string, entryKey string, entryOp string) (string, error) {
	// http://host:port/token?entryKey=$key&entryOp=put
	tokenURL := fmt.Sprintf("%s/token?entryKey=%s&entryOp=%s", serverURL, entryKey, entryOp)
	log.Println(tokenURL)
	tokenResp, err := http.Post(tokenURL, "application/json", strings.NewReader(""))
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	defer tokenResp.Body.Close()
	tokenBody, err := ioutil.ReadAll(tokenResp.Body)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	log.Println(string(tokenBody))
	token := tokenStruct{}
	err = json.Unmarshal(tokenBody, &token)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	log.Printf("put token: %s", token.Token)
	return token.Token, nil
}

func upload(serverURL, entryKey, token, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()

	// http://host:port/pblocks/$key?token=$ptoken Content-Type: application/octet-stream
	postURL := fmt.Sprintf("%s/pblocks/%s?token=%s", serverURL, entryKey, token)
	postResp, err := http.Post(postURL, "application/octet-stream", file)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer postResp.Body.Close()
	postBody, err := ioutil.ReadAll(postResp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(postBody))
	return
}

func download(serverURL, entryKey, token, filename string) {
	// http://host:port/pblocks/$key?token=$token
	getURL := fmt.Sprintf("%s/pblocks/%s?token=%s", serverURL, entryKey, token)
	getResp, err := http.Get(getURL)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer getResp.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()
	io.Copy(file, getResp.Body)

	log.Println("Download finish ", filename)
	return
}

func md5sum(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	md5h := md5.New()
	io.Copy(md5h, file)
	fmd5 := fmt.Sprintf("%x", md5h.Sum([]byte("")))
	return fmd5, nil
}

func onlyUpload(serverURL, key, filename string) {
	randKey := fmt.Sprintf("%s-sufix", key)
	ptoken, err := token(serverURL, randKey, "put")
	if err != nil {
		log.Fatal(err)
	}
	upload(serverURL, randKey, ptoken, filename)
	umd5, err := md5sum(filename)
	if err != nil {
		log.Println("calc ufile md5 error: ", err)
	}
	fmt.Printf("file: %s, md5: %s", filename, umd5)
}

func onlyDownload(serverURL, key, filename string) {
	gtoken, err := token(serverURL, key, "get")
	if err != nil {
		log.Fatal(err)
		return
	}
	var sfile string
	if filename == "filename" {
		sfile = fmt.Sprintf("file-%s.down", key)
	} else {
		sfile = filename
	}
	download(serverURL, key, gtoken, sfile)
	dmd5, err := md5sum(sfile)
	if err != nil {
		log.Println("calc sfile md5 error: ", err)
	}
	fmt.Printf("file: %s, md5: %s", filename, dmd5)
}

func validateUploadDownload(serverURL string, key string, num uint) string {
	rand.Seed(time.Now().Unix())
	randValue := rand.Intn(8192)

	ufile := fmt.Sprintf("ufile.%05d", randValue)
	file, err := os.OpenFile(ufile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal("OpenFile", err)
		return "error"
	}
	defer file.Close()
	var i uint
	for i = 0; i < num; i++ {
		randKey := fmt.Sprintf("%s-%04d-%05d", key, randValue, i)
		if _, err := file.Write([]byte("Append some data")); err != nil {
			log.Fatal("Write file")
		}
		ptoken, err := token(serverURL, randKey, "put")
		if err != nil {
			log.Fatal(err)
		}
		upload(serverURL, randKey, ptoken, ufile)
		umd5, err := md5sum(ufile)
		if err != nil {
			log.Println("calc ufile md5 error: ", err)
			break
		}

		gtoken, err := token(serverURL, randKey, "get")
		if err != nil {
			log.Fatal(err)
			break
		}

		sfile := fmt.Sprintf("sfile.%d", i)
		download(serverURL, randKey, gtoken, sfile)
		dmd5, err := md5sum(sfile)
		if err != nil {
			log.Println("calc sfile md5 error: ", err)
			break
		}
		if umd5 != dmd5 {
			log.Printf("checkmd5 failed  %s != %s", umd5, dmd5)
		} else {
			log.Printf("checkmd5 success %s == %s", umd5, dmd5)
			os.Remove(sfile)
		}

	}
	return "finish"
}

func main() {
	host := flag.String("host", "localhost", "hostname")
	port := flag.Int("port", 3000, "port")
	key := flag.String("key", "key", "key")
	ufile := flag.String("ufile", "filename", "upload filename")
	sfile := flag.String("save", "filename", "save download filename")
	num := flag.Uint("num", 1024, "upload/download times")
	debug := flag.Bool("debug", false, "enable/disable debug mode")
	ignore := flag.Bool("i", false, "ignore failed validation")

	flag.Parse()

	serverURL := fmt.Sprintf("http://%s:%d", *host, *port)
	fmt.Printf("key:%s, ufile:%s, dfile:%s debug: %v, num:%d, ignore:%v\n", *key, *ufile, *sfile, *debug, *num, *ignore)
	log.Printf("key:%s, ufile:%s, dfile:%s debug: %v, num:%d, ignore:%v\n", *key, *ufile, *sfile, *debug, *num, *ignore)

	if *ufile != "filename" {
		onlyUpload(serverURL, *key, *ufile)
	} else if *key != "key" {
		onlyDownload(serverURL, *key, *sfile)
	} else {
		validateUploadDownload(serverURL, *key, *num)
	}

}

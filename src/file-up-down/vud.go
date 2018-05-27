package main

/*
 * A commandline tool to upload/download and validate file
 * 2018-05-27
 */

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
	"path/filepath"
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
		fmt.Println(err)
		log.Fatal(err)
		return "", err
	}

	defer tokenResp.Body.Close()
	tokenBody, err := ioutil.ReadAll(tokenResp.Body)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	log.Printf("token: %s, statue: %s", string(tokenBody), tokenResp.Status)
	token := tokenStruct{}
	err = json.Unmarshal(tokenBody, &token)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

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
	log.Printf("uploadFinish: %s, status: %s", string(postBody), postResp.Status)
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
	log.Println("Download finish ", filename, getResp.Status)
	log.Printf("downloadFinish: %s, status: %s", filename, getResp.Status)
	return
}

func md5sum(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer file.Close()

	md5h := md5.New()
	io.Copy(md5h, file)
	fmd5 := fmt.Sprintf("%x", md5h.Sum([]byte("")))
	return fmd5, nil
}

func onlyUpload(serverURL, key, filename string) {
	md5, err := md5sum(filename)
	if err != nil {
		log.Fatal(err)
		return
	}
	fileKey := fmt.Sprintf("%s-%s", key, md5)
	ptoken, err := token(serverURL, fileKey, "put")
	if err != nil {
		log.Fatal(err)
		return
	}
	upload(serverURL, fileKey, ptoken, filename)

	fmt.Printf("file: %s \nkey : %s \nmd5 : %s\n", filename, fileKey, md5)
}

func onlyDownload(serverURL, key, dir, filename string) {
	gtoken, err := token(serverURL, key, "get")
	if err != nil {
		log.Fatal(err)
		return
	}
	var dfile string
	if filename == "filename" {
		dfile = filepath.Join(dir, fmt.Sprintf("file-%s.down", key))
	} else {
		dfile = filename
	}
	download(serverURL, key, gtoken, dfile)
	md5, err := md5sum(dfile)
	if err != nil {
		log.Println("calc sfile md5 error: ", err)
	}

	fmt.Printf("file: %s \nkey : %s \nmd5 : %s\n", dfile, key, md5)
}

func validateUploadDownload(serverURL string, key string, dir string, num uint) {

	ufile := filepath.Join(dir, fmt.Sprintf("upload-file.%s", key))

	file, err := os.OpenFile(ufile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal("OpenFile", err)
		return
	}
	defer file.Close()

	var i uint
	for i = 0; i < num; i++ {
		randKey := fmt.Sprintf("%s-%08d", key, i)
		content := fmt.Sprintf("Just test file contents with( %s )\n", randKey)
		if _, err := file.Write([]byte(content)); err != nil {
			log.Fatal("Write file", err)
			break
		}

		ptoken, err := token(serverURL, randKey, "put")
		if err != nil {
			log.Fatal(err)
			break
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

		dfile := filepath.Join(dir, fmt.Sprintf("download-file-%s.%d", key, i))
		download(serverURL, randKey, gtoken, dfile)

		dmd5, err := md5sum(dfile)
		if err != nil {
			log.Println("calc download file md5 error: ", err)
			break
		}
		if umd5 != dmd5 {
			log.Printf("checkmd5 %s failed  %s != %s", dfile, umd5, dmd5)
		} else {
			log.Printf("checkmd5 %s success %s == %s", dfile, umd5, dmd5)
			os.Remove(dfile)
		}

	}

}

func main() {
	const VERSION = "version: 1.0.0"
	host := flag.String("host", "localhost", "hostname")
	port := flag.Int("port", 3000, "port")
	key := flag.String("key", "key", "key")
	ufile := flag.String("ufile", "filename", "upload filename")
	dfile := flag.String("dfile", "filename", "download filename")
	num := flag.Uint("num", 1, "upload/download times")
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
	datadir := filepath.Join(pwd, "tmp")
	_ = os.Mkdir(logdir, 0755)
	_ = os.Mkdir(datadir, 0755)

	rand.Seed(time.Now().Unix())
	randValue := rand.Intn(8192)

	logFilename := fmt.Sprintf("%s-%04d.log", os.Args[0], randValue)
	logFilename = filepath.Join(logdir, logFilename)
	logFile, logErr := os.OpenFile(logFilename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if logErr != nil {
		fmt.Println("Fail to OpenFile", logErr)
		os.Exit(1)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	serverURL := fmt.Sprintf("http://%s:%d", *host, *port)
	fmt.Printf("server:%s, logfile:%s\n", serverURL, logFilename)
	log.Printf("server:%s, logfile:%s debug: %v, num:%d, ignore:%v\n", serverURL, logFilename, *debug, *num, *ignore)

	if *ufile != "filename" {
		onlyUpload(serverURL, *key, *ufile)
	} else if *key != "key" {
		onlyDownload(serverURL, *key, datadir, *dfile)
	} else {
		randKey := fmt.Sprintf("%s-%04d", *key, randValue)
		validateUploadDownload(serverURL, randKey, datadir, *num)
	}

}

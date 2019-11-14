package main

/*
 * A commandline tool to upload/download and validate file
 * 2018-05-27T09:09:00Z
 */

import (
	"crypto/md5"
	"encoding/json"
	"errors"
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

// BuildDate to record build date
var BuildDate = "Unknow Date"

// Version to record build bersion
var Version = "1.0.3"

type tokenStruct struct {
	Token string
}

func token(serverURL string, entryKey string, entryOp string) (string, error) {
	// http://host:port/token?entryKey=$key&entryOp=put
	tokenURL := fmt.Sprintf("%s/token?entryKey=%s&entryOp=%s", serverURL, entryKey, entryOp)
	log.Println(tokenURL)
	/* Maybe need a timeout control
	timeout := time.Duration(60 * time.Second)
	client := &http.Client{Timeout: timeout}
	*/
	tokenResp, err := http.Post(tokenURL, "application/json", strings.NewReader(""))
	if err != nil {
		fmt.Println(err)
		log.Println(err)
		return "", err
	}

	tokenResp.Close = true
	defer tokenResp.Body.Close()
	if tokenResp.StatusCode != 200 {
		log.Println("Server error: ", tokenResp.Status)
		return "", errors.New(tokenResp.Status)
	}

	tokenBody, err := ioutil.ReadAll(tokenResp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}

	log.Printf("token: %s, statue: %s", string(tokenBody), tokenResp.Status)
	token := tokenStruct{}
	err = json.Unmarshal(tokenBody, &token)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return token.Token, nil
}

func upload(serverURL, entryKey, token, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		log.Println(err)
		return err
	}
	defer file.Close()

	// http://host:port/pblocks/$key?token=$ptoken Content-Type: application/octet-stream
	postURL := fmt.Sprintf("%s/pblocks/%s?token=%s", serverURL, entryKey, token)
	postResp, err := http.Post(postURL, "application/octet-stream", file)
	if err != nil {
		log.Println(err)
		return err
	}

	postResp.Close = true
	defer postResp.Body.Close()
	if postResp.StatusCode != 200 {
		log.Println("Server error: ", postResp.Status)
		return errors.New(postResp.Status)
	}
	postBody, err := ioutil.ReadAll(postResp.Body)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("uploadFinish: %s, status: %s", string(postBody), postResp.Status)
	return nil
}

func download(serverURL, entryKey, token, filename string) error {
	// http://host:port/pblocks/$key?token=$token
	getURL := fmt.Sprintf("%s/pblocks/%s?token=%s", serverURL, entryKey, token)
	getResp, err := http.Get(getURL)
	if err != nil {
		fmt.Println(err)
		log.Println(err)
		return err
	}
	getResp.Close = true
	defer getResp.Body.Close()
	if getResp.StatusCode != 200 {
		log.Println("Server error: ", getResp.Status)
		return errors.New(getResp.Status)
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	io.Copy(file, getResp.Body)
	log.Printf("downloadFinish: %s, status: %s", filename, getResp.Status)
	return nil
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
	if err := upload(serverURL, fileKey, ptoken, filename); err != nil {
		log.Println("upload error: ", err)
		fmt.Println("upload error: ", err)
	} else {
		fmt.Printf("file: %s \nkey : %s \nmd5 : %s\n", filename, fileKey, md5)
	}

}

func onlyDownload(serverURL, key, dir, filename string) {
	gtoken, err := token(serverURL, key, "get")
	if err != nil {
		log.Fatal(err)
		return
	}
	var dfile string
	if filename == "" {
		dfile = filepath.Join(dir, fmt.Sprintf("download-%s", key))
	} else {
		dfile = filename
	}
	if err := download(serverURL, key, gtoken, dfile); err != nil {
		log.Println("download error: ", err)
		fmt.Println("download error: ", err)
	} else {
		md5, err := md5sum(dfile)
		if err != nil {
			log.Println("calc sfile md5 error: ", err)
		}

		fmt.Printf("file: %s \nkey : %s \nmd5 : %s\n", dfile, key, md5)
	}
}

func validateUploadDownload(serverURL string, key string, dir string, num uint, ignore bool) {

	ufile := filepath.Join(dir, fmt.Sprintf("upload-%s", key))

	file, err := os.OpenFile(ufile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal("OpenFile", err)
		return
	}
	defer file.Close()

	var i uint
	var totalSize, fileSize uint = 0, 0
	for i = 0; i < num; i++ {
		randKey := fmt.Sprintf("%s-%09d", key, i)
		content := fmt.Sprintf("Just tests data injection with contents ( %s )\n", randKey)
		if _, err := file.Write([]byte(content)); err != nil {
			log.Println("Write file", err)
			fmt.Println("Write file", err)
			if ignore {
				continue
			} else {
				break
			}
		}
		fileSize += uint(len(content))

		ptoken, err := token(serverURL, randKey, "put")
		if err != nil {
			fmt.Println("put token: ", err)
			log.Println("put token: ", err)
			if ignore {
				continue
			} else {
				break
			}
		}

		if err := upload(serverURL, randKey, ptoken, ufile); err != nil {
			fmt.Println("upload error: ", err)
			log.Println("upload error: ", err)
			if ignore {
				continue
			} else {
				break
			}
		}
		totalSize += fileSize
		umd5, err := md5sum(ufile)
		if err != nil {
			fmt.Println("calc ufile md5 error: ", err)
			log.Println("calc ufile md5 error: ", err)
			if ignore {
				continue
			} else {
				break
			}
		}

		gtoken, err := token(serverURL, randKey, "get")
		if err != nil {
			fmt.Println("get token: ", err)
			log.Println("get token: ", err)
			if ignore {
				continue
			} else {
				break
			}
		}

		dname := fmt.Sprintf("download-%s", randKey)
		dfile := filepath.Join(dir, dname)
		if err := download(serverURL, randKey, gtoken, dfile); err != nil {
			fmt.Println("download error: ", err)
			log.Println("download error: ", err)
			if ignore {
				continue
			} else {
				break
			}
		}

		dmd5, err := md5sum(dfile)
		if err != nil {
			fmt.Println("calc download file md5 error: ", err)
			log.Println("calc download file md5 error: ", err)
			if ignore {
				continue
			} else {
				break
			}
		}

		if umd5 != dmd5 {
			log.Printf("checkmd5 %s failed  %s != %s", dname, umd5, dmd5)
			fmt.Printf("checkmd5 %s failed  %s != %s\n", dname, umd5, dmd5)
			if ignore {
				continue
			} else {
				break
			}
		} else {
			log.Printf("check %s md5 %s success, filesize: %d, totalSize: %d", dname, dmd5, fileSize, totalSize)
			fmt.Printf("check %s md5 %s success, filesize: %d, totalSize: %d\n", dname, dmd5, fileSize, totalSize)
			os.Remove(dfile)
		}

	}

}

func main() {
	var VERSION = fmt.Sprintf("Version: %s  build: %s", Version, BuildDate)
	host := flag.String("host", "127.0.0.1", "server address")
	port := flag.Int("port", 3000, "server port")
	key := flag.String("key", "key", "file key")
	ufile := flag.String("ufile", "", "upload filename")
	dfile := flag.String("dfile", "", "download save filename")
	num := flag.Uint("num", 1, "upload/download times")
	ignore := flag.Bool("i", false, "ignore failed validation")
	version := flag.Bool("version", false, "show version")

	flag.Parse()
	if *version == true {
		fmt.Printf("%s  %s\n", filepath.Base(os.Args[0]), VERSION)
		return
	}

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	logdir := filepath.Join(pwd, "log")
	datadir := filepath.Join(pwd, "tmp")
	_ = os.MkdirAll(logdir, 0755)
	_ = os.MkdirAll(datadir, 0755)

	rand.Seed(time.Now().Unix())
	randValue := rand.Intn(99999)

	logFilename := fmt.Sprintf("%s-%05d.log", filepath.Base(os.Args[0]), randValue)
	logFilename = filepath.Join(logdir, logFilename)
	logFile, logErr := os.OpenFile(logFilename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if logErr != nil {
		fmt.Println("Fail to OpenFile", logErr)
		os.Exit(1)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	serverURL := fmt.Sprintf("http://%s:%d", *host, *port)
	fmt.Printf("server: %s, logfile: %s\n", serverURL, logFilename)
	log.Printf("server: %s, logfile: %s, num: %d, ignore: %v\n", serverURL, logFilename, *num, *ignore)

	if *ufile != "" {
		onlyUpload(serverURL, *key, *ufile)
	} else if *key != "key" {
		onlyDownload(serverURL, *key, datadir, *dfile)
	} else {
		randKey := fmt.Sprintf("%s-%05d", *key, randValue)
		validateUploadDownload(serverURL, randKey, datadir, *num, *ignore)
	}

}

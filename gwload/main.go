package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	mrand "math/rand"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"code.cloudfoundry.org/bytefmt"
)

var (
	gw         string
	sc         string
	appID      string
	appKey     string
	urlType    string
	uinfo      string
	endpoint   string
	bucket     string
	accessKey  string
	secretKey  string
	maxSizeArg string
	minSizeArg string
	concurent  int
	rounds     int
	debug      bool
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxIdleConnsPerHost: 10,
		DisableCompression:  false,
		DisableKeepAlives:   false,
	},
}

func init() {
	mrand.Seed(time.Now().UnixNano())
}

// RandomString gen random string with len
func RandomString(len int) string {
	buf := make([]byte, len)
	_, err := rand.Read(buf)
	if err != nil {
		for i := 0; i < len; i++ {
			buf[i] = byte(mrand.Intn(128))
		}
	}
	return hex.EncodeToString(buf)
}

func presignV2(method, endpoint, bucket, key, contentType, ak, sk string, exp int64) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	expStr := strconv.FormatInt(time.Now().Unix()+exp, 10)

	q := u.Query()
	q.Set("AWSAccessKeyId", ak)
	q.Set("Expires", expStr)
	u.Path = fmt.Sprintf("/%s/%s", bucket, key)

	contentMd5 := "" // header Content-MD5
	strToSign := fmt.Sprintf("%s\n%s\n%s\n%v\n%s", method, contentMd5, contentType, expStr, u.EscapedPath())

	mac := hmac.New(sha1.New, []byte(sk))
	mac.Write([]byte(strToSign))

	q.Set("Signature", base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	u.RawQuery = q.Encode()

	return u.String(), nil
}
func gwSignatue(key, uinfo, URL string, exp int64, extra []string) string {
	buffer := bytes.NewBufferString(URL)
	buffer.WriteString("\n")
	buffer.WriteString(strconv.FormatInt(exp, 10))
	if uinfo != "" {
		buffer.WriteString("\n")
		buffer.WriteString(uinfo)
	}

	for _, v := range extra {
		buffer.WriteString("\n")
		buffer.WriteString(v)
	}
	//log.Printf("gwSignatue: %s\n", buffer.Bytes())
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write(buffer.Bytes())
	return base64.URLEncoding.EncodeToString(mac.Sum(nil))
}

//GenStaticURL generate static URL
func GenStaticURL(gw, appID, appKey, presignURL string) (string, error) {
	if appID == "" {
		return "", errors.New("invalid appID")
	}
	if appKey == "" {
		return "", errors.New("invalid appKey")
	}
	gwURL, err := url.ParseRequestURI(gw)
	if err != nil {
		return "", fmt.Errorf("invalid gateway addr, %s", err)
	}
	psgURL, err := url.ParseRequestURI(presignURL)
	if err != nil {
		return "", fmt.Errorf("invalid presignURL, %s", err)
	}
	sign := gwSignatue(appKey, "", psgURL.RequestURI(), 0, nil)

	return fmt.Sprintf("%s/s/%s/%s/%s/%d",
		gwURL.String(),
		base64.URLEncoding.EncodeToString([]byte(psgURL.RequestURI())),
		sign,
		appID,
		0), nil
}

// GenDynamic2URL generate dynamic2 URL
func GenDynamic2URL(gw, appID, appKey, presignURL, uinfo string, exp int64) (string, error) {
	if appID == "" {
		return "", errors.New("invalid appID")
	}
	if appKey == "" {
		return "", errors.New("invalid appKey")
	}
	if uinfo == "" {
		return "", errors.New("invalid cmb_uinfo")
	}
	if exp <= time.Now().Unix() {
		return "", errors.New("invalid exp timestamp")
	}
	gwURL, err := url.ParseRequestURI(gw)
	if err != nil {
		return "", fmt.Errorf("invalid gateway addr, %s", err)
	}
	psgURL, err := url.ParseRequestURI(presignURL)
	if err != nil {
		return "", fmt.Errorf("invalid presignURL, %s", err)
	}

	sign := gwSignatue(appKey, uinfo, psgURL.RequestURI(), exp, nil)

	return fmt.Sprintf("%s/g/%s/%s/%s/%d",
		gwURL.String(),
		base64.URLEncoding.EncodeToString([]byte(psgURL.RequestURI())),
		sign,
		appID,
		exp), nil
}

func objectSize(min, max uint64) uint64 {
	if min >= max {
		return max
	}
	return mrand.Uint64()%(max-min) + min
}

func prettyRequest(req *http.Request) (s string) {
	s = fmt.Sprintf("\nMethod: %s\n", req.Method)
	s = fmt.Sprintf("%sURL: %s\n", s, req.URL.String())
	s = fmt.Sprintf("%sContent-Length: %v\n", s, req.ContentLength)
	s = fmt.Sprintf("%sHost: %v\n", s, req.Host)
	for k, v := range req.Header {
		s = fmt.Sprintf("%s  %v: %v\n", s, k, v)
	}
	return
}

func prettyResponse(resp *http.Response) (s string) {
	s = fmt.Sprintf("\nContent-Length: %v\n", resp.ContentLength)
	for k, v := range resp.Header {
		s = fmt.Sprintf("%s  %v: %v\n", s, k, v)
	}
	if resp.Body != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		s = fmt.Sprintf("%sbody:\n%s\n", s, body)
	}
	return
}

func main() {
	flag.StringVar(&gw, "gw", "", "GW address")
	//flag.StringVar(&sc, "sc", "", "sc address")
	flag.StringVar(&accessKey, "ak", "", "S3 access key")
	flag.StringVar(&secretKey, "sk", "", "S3 secret key")
	flag.StringVar(&appID, "i", "", "App ID")
	flag.StringVar(&appKey, "k", "", "App key")
	flag.StringVar(&uinfo, "u", "uinfo-value", "UInfo")
	flag.StringVar(&urlType, "t", "static", "GW url type(static,dyanmic2)")
	flag.StringVar(&endpoint, "e", "", "S3 endpoint")
	flag.StringVar(&bucket, "b", "", "Bucket name")
	flag.IntVar(&rounds, "n", 1, "Number of rounds to run")
	flag.IntVar(&concurent, "c", 20, "Number of requests to run concurrently")
	flag.StringVar(&maxSizeArg, "max", "10M", "Max size of objects in bytes with postfix K, M, and G")
	flag.StringVar(&minSizeArg, "min", "2M", "Min size of objects in bytes with postfix K, M, and G")
	flag.BoolVar(&debug, "debug", false, "debug log level")
	flag.Parse()
	if gw == "" || endpoint == "" {
		fmt.Printf("unknown gw:%v, endpoint:%v\n", gw, endpoint)
		flag.Usage()
		return
	}

	if urlType != "static" && urlType != "dynamic2" {
		fmt.Printf("unknown GW URL Type:%v\n", urlType)
		flag.Usage()
		return
	}

	maxObjSize, err := bytefmt.ToBytes(maxSizeArg)
	if err != nil {
		log.Fatalf("Invalid -max argument for object size: %v", err)
	}
	minObjSize, err := bytefmt.ToBytes(minSizeArg)
	if err != nil {
		log.Fatalf("Invalid -min argument for object size: %v", err)
	}
	if minObjSize > maxObjSize {
		log.Fatalf("Invalid -min argument for object size: %v", err)
	}

	gwURL, err := url.Parse(gw)
	if err != nil {
		log.Fatalf("Invalid gw URL: %v", err)
	}

	var totalUploadCount int32
	var totalUploadFailedCount int32

	objectData := make([]byte, maxObjSize)
	if n, e := rand.Read(objectData); e != nil {
		log.Fatalf("generate random data failed: %s", e)
	} else if uint64(n) < maxObjSize {
		log.Fatalf("invalid random data size, got %d, expect %d", n, maxObjSize)
	}

	for r := 1; r <= rounds; r++ {
		var uploadCount, uploadFailedCount int32
		wg := sync.WaitGroup{}
		for n := 1; n <= concurent; n++ {
			wg.Add(1)
			go func() {
				atomic.AddInt32(&uploadCount, 1)
				randomSize := objectSize(minObjSize, maxObjSize)
				fileobj := bytes.NewReader(objectData[0:randomSize])
				key := RandomString(18)
				presignURL, err := presignV2(http.MethodPut, endpoint, bucket, key, "application/octet-stream", accessKey, secretKey, 684000)
				if err != nil {
					log.Fatal("presign: ", err, gwURL)
					return
				}
				var gwURL string

				if urlType == "dynamic2" {
					gwURL, err = GenDynamic2URL(gw, appID, appKey, presignURL, uinfo, time.Now().Add(1*time.Hour).Unix())
					if err != nil {
						log.Fatal("GenGWURL: ", presignURL, err)
						return
					}
				} else if urlType == "static" {
					gwURL, err = GenStaticURL(gw, appID, appKey, presignURL)
					if err != nil {
						log.Fatal("GenGWURL: ", presignURL, err)
						return
					}
					uinfo = ""
				}

				req, err := http.NewRequest(http.MethodPut, gwURL, fileobj)
				if err != nil {
					log.Fatal("NewRequest: ", err)
					return
				}

				//req.Header.Set("Content-Length", strconv.FormatUint(randomSize, 10))
				req.Header.Set("Content-Type", "application/octet-stream")
				if uinfo != "" {
					req.Header.Set("Cmb_uinfo", uinfo)
				}

				if debug {
					trace := &httptrace.ClientTrace{
						DNSStart: func(info httptrace.DNSStartInfo) {
							fmt.Printf("DNS start %v for %v\n", time.Now(), info.Host)
						},
						DNSDone: func(info httptrace.DNSDoneInfo) {
							fmt.Printf("DNS start %v for %v\n", time.Now(), info.Addrs)
						},
					}
					req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
				}
				resp, err := httpClient.Do(req)
				if err != nil {
					if resp != nil {
						log.Fatalf("FATAL: Error uploading object: error: %s\n%s\n%s\n", err, prettyRequest(req), prettyResponse(resp))
					} else {
						log.Fatalf("FATAL: Error uploading object: error: %s\n%s\n", err, prettyRequest(req))
					}
				} else if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
					atomic.AddInt32(&uploadFailedCount, 1)
					atomic.AddInt32(&uploadCount, -1)
				}
				msg := ""
				if debug {
					msg = fmt.Sprintf("%v: %s<==============>%s", resp.StatusCode, prettyRequest(req), prettyResponse(resp))
				} else {
					msg = fmt.Sprintf("%v: %s", resp.StatusCode, gwURL)
				}
				fmt.Println(msg)
				if resp.Body != nil {
					resp.Body.Close()
				}
				wg.Done()
			}()
		}
		wg.Wait()

		totalUploadCount += uploadCount
		totalUploadFailedCount += totalUploadFailedCount
		fmt.Printf("%4d\t\t%v/%v\n", r, uploadFailedCount, uploadCount)
	}

	fmt.Printf("done\t\t%v/%v\n", totalUploadFailedCount, totalUploadCount)
}

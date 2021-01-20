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
	"strings"
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
	idPrefix   string
	keyPrefix  string
	concurent  int
	timeout    uint
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

func indexN(s string, c byte, n int) int {
	length := len(s)
	count := 0
	for i := 0; i < length; i++ {
		if s[i] == c {
			count++
		}
		if count == n {
			return i
		}
	}
	return -1
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

// GenStaticV2URL generate static v2 URL
// /s2/{app}/{sign}/{bucket:[0-9A-Za-z_.-]{3,64}}/{key:.*}?ak=xxx&signature=ssss
func GenStaticV2URL(gw, app, key, presignURL string) (string, error) {
	if app == "" {
		return "", errors.New("invalid appID")
	}
	if key == "" {
		return "", errors.New("invalid appKey")
	}
	if strings.HasPrefix(gw, "http") == false {
		return "", errors.New("invalid gateway addr")
	}
	if strings.HasPrefix(presignURL, "http") == false {
		return "", errors.New("invalid presignURL")
	}
	psgURL, err := url.ParseRequestURI(presignURL)
	if err != nil {
		return "", err
	}

	sign := gwSignatue(key, "", psgURL.RequestURI(), 0, nil)

	// server/s2/{app}/{sign}/{bucket:[0-9A-Za-z_.-]{3,64}}/{key:.*}?ak=xxx&signature=ssss
	pos := indexN(presignURL, '/', 3)
	if pos < 0 {
		return "", errors.New("invalid presignURL")
	}

	return fmt.Sprintf("%s/s2/%s/%s%s", gw, app, sign, presignURL[pos:]), nil
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
	s = fmt.Sprintf("%s\n", req.URL.String())
	s = fmt.Sprintf("%s  Content-Length: %v\n", s, req.ContentLength)
	s = fmt.Sprintf("%s  Host: %v\n", s, req.Host)
	for k, v := range req.Header {
		s = fmt.Sprintf("%s  %v: %v\n", s, k, v)
	}
	return
}

func prettyResponse(resp *http.Response) (s string) {
	s = fmt.Sprintf("\n  Content-Length: %v\n", resp.ContentLength)
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
	flag.StringVar(&accessKey, "ak", "gatewaybizverify", "S3 access key")
	flag.StringVar(&secretKey, "sk", "", "S3 secret key")
	flag.StringVar(&appID, "i", "gatewaybizverify", "App ID")
	flag.StringVar(&appKey, "k", "", "App key")
	flag.StringVar(&uinfo, "u", "uinfo-value", "UInfo")
	flag.StringVar(&urlType, "t", "static", "GW url type(static,staticv2,dyanmic2)")
	flag.StringVar(&endpoint, "e", "http://ecsnp01.s3ecs.itcenter.cmbchina.cn:9020", "S3 endpoint")
	flag.StringVar(&bucket, "b", "", "Bucket name")
	flag.IntVar(&concurent, "c", 100, "Number of requests to run concurrently")
	flag.UintVar(&timeout, "T", 300, "Timeout for each request in seconds")
	flag.StringVar(&maxSizeArg, "max", "10M", "Max size of objects in bytes with postfix K, M, and G")
	flag.StringVar(&minSizeArg, "min", "2M", "Min size of objects in bytes with postfix K, M, and G")
	flag.StringVar(&idPrefix, "id-prefix", "i001", "Prefix of header x-request-id")
	flag.StringVar(&keyPrefix, "key-prefix", "k001", "Prefix of Object name")
	flag.BoolVar(&debug, "debug", false, "debug log level")
	flag.Parse()
	if gw == "" || endpoint == "" {
		fmt.Printf("unknown gw:%v, endpoint:%v\n", gw, endpoint)
		flag.Usage()
		return
	}

	if urlType != "static" && urlType != "staticv2" && urlType != "dynamic2" {
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

	objectData := make([]byte, maxObjSize)
	if n, e := rand.Read(objectData); e != nil {
		log.Fatalf("generate random data failed: %s", e)
	} else if uint64(n) < maxObjSize {
		log.Fatalf("invalid random data size, got %d, expect %d", n, maxObjSize)
	}

	httpClient.Timeout = time.Duration(timeout) * time.Second

	beginTime := time.Now()
	fmt.Println("begin: ", beginTime)
	var uploadFailedCount int32
	wg := sync.WaitGroup{}
	for n := 1; n <= concurent; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			randomSize := objectSize(minObjSize, maxObjSize)
			fileobj := bytes.NewReader(objectData[0:randomSize])
			randomStr := RandomString(10)

			key := keyPrefix + "-" + randomStr
			presignURL, err := presignV2(http.MethodPut, endpoint, bucket, key, "application/octet-stream", accessKey, secretKey, 684000)
			if err != nil {
				atomic.AddInt32(&uploadFailedCount, 1)
				log.Fatal("presign: ", err, gwURL)
				return
			}

			var gwURL string
			if urlType == "dynamic2" {
				gwURL, err = GenDynamic2URL(gw, appID, appKey, presignURL, uinfo, time.Now().Add(1*time.Hour).Unix())
				if err != nil {
					atomic.AddInt32(&uploadFailedCount, 1)
					log.Fatal("Gen dynamic2 URL error: ", presignURL, err)
					return
				}
			} else if urlType == "static" {
				gwURL, err = GenStaticURL(gw, appID, appKey, presignURL)
				if err != nil {
					atomic.AddInt32(&uploadFailedCount, 1)
					log.Fatal("Gen static URL error: ", presignURL, err)
					return
				}
				uinfo = ""
			} else if urlType == "staticv2" {
				gwURL, err = GenStaticV2URL(gw, appID, appKey, presignURL)
				if err != nil {
					log.Fatal("Gen staticv2 URL error: ", presignURL, err)
					return
				}
				uinfo = ""
			} else {
				atomic.AddInt32(&uploadFailedCount, 1)
				log.Fatal("unknown URL type: ", urlType)
				return
			}

			req, err := http.NewRequest(http.MethodPut, gwURL, fileobj)
			if err != nil {
				atomic.AddInt32(&uploadFailedCount, 1)
				log.Fatal("NewRequest error: ", err)
				return
			}

			//req.Header.Set("Content-Length", strconv.FormatUint(randomSize, 10))
			req.Header.Set("X-request-id", fmt.Sprintf("%s-%s-%v", idPrefix, randomStr, n))
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
			startTime := time.Now()
			resp, err := httpClient.Do(req)
			if err != nil {
				atomic.AddInt32(&uploadFailedCount, 1)
				if resp != nil {
					log.Printf("uploading Object error: %s\n%s\n%s\n", err, prettyRequest(req), prettyResponse(resp))
				} else {
					log.Printf("uploading Object error: %s\n%s\n", err, prettyRequest(req))
				}
				return
			}

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				atomic.AddInt32(&uploadFailedCount, 1)
			}

			msg := fmt.Sprintf("%v %v", resp.StatusCode, time.Since(startTime).Milliseconds())
			if debug {
				msg = fmt.Sprintf("%s %s<==============>%s", msg, prettyRequest(req), prettyResponse(resp))
			} else {
				msg = fmt.Sprintf("%s %s", msg, gwURL)
			}
			fmt.Println(msg)

			if resp.Body != nil {
				resp.Body.Close()
			}
		}(n)
	}
	wg.Wait()
	endtime := time.Now()
	fmt.Printf("done\t%v/%v\t%v\t%s\n", uploadFailedCount, concurent, endtime.Sub(beginTime).Milliseconds(), endtime)
}

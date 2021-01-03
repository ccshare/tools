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
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"code.cloudfoundry.org/bytefmt"
)

var (
	gw        string
	sc        string
	appID     string
	appKey    string
	urlType   string
	endpoint  string
	bucket    string
	accessKey string
	secretKey string
	sizeArg   string
	concurent int
	rounds    int
)

func init() {
	mrand.Seed(time.Now().UnixNano())
}

// transport represent Our HTTP transport used for the roundtripper below
var transport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).Dial,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 0,
	// Allow an unlimited number of idle connections
	MaxIdleConnsPerHost: 4096,
	MaxIdleConns:        0,
	// But limit their idle time
	IdleConnTimeout: time.Minute,
	// Ignore TLS errors
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
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

func main() {
	flag.StringVar(&gw, "gw", "", "gw address")
	flag.StringVar(&sc, "sc", "", "sc address")
	flag.StringVar(&accessKey, "ak", "", "S3 access key")
	flag.StringVar(&secretKey, "sk", "", "S3 secret key")
	flag.StringVar(&appID, "i", "", "app ID")
	flag.StringVar(&appKey, "k", "", "app key")
	flag.StringVar(&urlType, "t", "", "url type(open,gwopen,static,dyanmic,dyanmic2,sec)")
	flag.StringVar(&endpoint, "e", "", "S3 endpoint")
	flag.StringVar(&bucket, "b", "", "bucket name")
	flag.IntVar(&rounds, "n", 10, "Number of rounds to run")
	flag.IntVar(&concurent, "c", 1, "Number of requests to run concurrently")
	flag.StringVar(&sizeArg, "z", "128K", "Size of objects in bytes with postfix K, M, and G")
	flag.Parse()
	if gw == "" || endpoint == "" {
		flag.Usage()
		return
	}
	httpClient := &http.Client{Transport: transport}
	var totalUploadTime float64
	var totalUploadCount int32
	var totalUploadFailedCount int32

	objectSize, err := bytefmt.ToBytes(sizeArg)
	if err != nil {
		log.Fatalf("Invalid -z argument for object size: %v", err)
	}
	fmt.Println("size: ", objectSize)

	objectData := make([]byte, objectSize)
	if n, e := rand.Read(objectData); e != nil {
		log.Fatalf("generate random data failed: %s", e)
	} else if uint64(n) < objectSize {
		log.Fatalf("invalid random data size, got %d, expect %d", n, objectSize)
	}

	// Loop running the tests
	fmt.Println("Loop\tMethod\t  Objects\tElapsed(s)\t Throuphput\t   TPS\t Failed")
	for r := 1; r <= rounds; r++ {
		var uploadCount, uploadFailedCount int32

		var uploadFinish time.Time
		starttime := time.Now()

		wg := sync.WaitGroup{}
		for n := 1; n <= concurent; n++ {
			wg.Add(1)
			go func() {
				fileobj := bytes.NewReader(objectData)
				key := RandomString(18)
				presignURL, err := presignV2(http.MethodPut, endpoint, bucket, key, "application/octet-stream", accessKey, secretKey, 684000)
				if err != nil {
					log.Fatal("presign: ", err)
					return
				}
				gwURL, err := GenStaticURL(gw, appID, appKey, presignURL)
				if err != nil {
					log.Fatal("GenGWURL: ", presignURL, err)
					return
				}
				fmt.Println("URL: ", presignURL, gwURL)
				req, err := http.NewRequest(http.MethodPut, gwURL, fileobj)
				if err != nil {
					log.Fatal("NewRequest: ", err)
					return
				}
				req.Header.Set("Content-Length", strconv.FormatUint(objectSize, 10))
				req.Header.Set("Content-Type", "application/octet-stream")

				if resp, err := httpClient.Do(req); err != nil {
					log.Fatalf("FATAL: Error uploading object %s: %v", presignURL, err)
				} else if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
					atomic.AddInt32(&uploadFailedCount, 1)
					atomic.AddInt32(&uploadCount, -1)
					fmt.Printf("upload resp: %v\n", resp)
					if resp.StatusCode != http.StatusServiceUnavailable {
						fmt.Printf("Upload status %s: resp: %+v\n", resp.Status, resp)
						if resp.Body != nil {
							body, _ := ioutil.ReadAll(resp.Body)
							fmt.Printf("Body: %s\n", string(body))
						}
					}
				}
				wg.Done()
			}()
		}
		wg.Wait()
		uploadTime := uploadFinish.Sub(starttime).Seconds()
		totalUploadTime += uploadTime
		totalUploadCount += uploadCount
		totalUploadFailedCount += totalUploadFailedCount
		bps := float64(uint64(uploadCount)*objectSize) / uploadTime
		fmt.Println(fmt.Sprintf("%4d\t%6s\t%9d\t%10.1f\t%10sB\t%6.1f\t%7d", r, http.MethodPut, uploadCount, uploadTime, bytefmt.ByteSize(uint64(bps)), float64(uploadCount)/uploadTime, uploadFailedCount))
	}
}

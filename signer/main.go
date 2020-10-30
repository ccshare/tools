package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

func main() {
	method := flag.String("m", "GET", "http method")
	server := flag.String("s", "http://192.168.55.2:9000/open/h0", "S3 request addr")
	//header := flag.String("H", "", "http header")
	ak := flag.String("ak", "object_user1", "access key")
	sk := flag.String("sk", "ChangeMeChangeMeChangeMeChangeMeChangeMe", "secret key")
	st := flag.String("t", "20201030T085240Z", "signing time")
	exp := flag.Int64("exp", 86400, "expire time seconds")
	flag.Parse()

	req, err := http.NewRequest(strings.ToUpper(*method), *server, nil)
	if err != nil {
		fmt.Println("new request error: ", err)
		return
	}

	cred := aws.Credentials{
		AccessKeyID:     *ak,
		SecretAccessKey: *sk,
		CanExpire:       true,
		Expires:         time.Now().Add(86400),
	}

	t, err := time.Parse("20060102T150405Z", *st)
	if err != nil {
		fmt.Println("parse signing time error: ", err)
		return
	}
	signer := v4.NewSigner()
	signer.DisableHeaderHoisting = true
	signer.DisableURIPathEscaping = true

	query := req.URL.Query()
	query.Set("X-Amz-Expires", strconv.FormatInt(*exp, 10))
	req.URL.RawQuery = query.Encode()

	u, h, err := signer.PresignHTTP(context.Background(), cred, req, "UNSIGNED-PAYLOAD", "s3", "default", t)
	if err != nil {
		fmt.Println("presign error: ", err)
		return
	}
	fmt.Println("presign:")
	for k, v := range h {
		fmt.Println(k, ":", v)
	}
	fmt.Println(u)

	err = signer.SignHTTP(context.Background(), cred, req, "", "s3", "default", t)
	if err != nil {
		fmt.Println("sign error: ", err)
		return
	}

	fmt.Println("sign:")
	for k, v := range req.Header {
		fmt.Println(k, ":", v)
	}
	fmt.Println(req.URL.String())
}

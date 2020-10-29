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
	server := flag.String("s", "http://127.0.0.1", "S3 server")
	header := flag.String("H", "", "http header")
	ak := flag.String("ak", "object_user1", "access key")
	sk := flag.String("sk", "ChangeMeChangeMeChangeMeChangeMeChangeMe", "secret key")
	//exp := flag.Int("exp", 86400, "expire time seconds")

	flag.Parse()

	fmt.Println("signer: ", *server, *header)

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

	t, err := time.Parse("20060102T150405Z", "20201029T073057Z")
	if err != nil {
		fmt.Println("parse time error: ", err)
		return
	}
	signer := v4.NewSigner()
	signer.DisableHeaderHoisting = true
	signer.DisableURIPathEscaping = true

	query := req.URL.Query()
	query.Set("X-Amz-Expires", strconv.FormatInt(86400, 10))
	req.URL.RawQuery = query.Encode()
	u, h, err := signer.PresignHTTP(context.Background(), cred, req, "", "s3", "default", t)
	if err != nil {
		fmt.Println("presign error: ", err)
		return
	}
	fmt.Println("presign url: ", u)
	fmt.Println("presign header: ", h)

}

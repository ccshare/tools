package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/spf13/cobra"
)

var (
	version string = "unknown"
)

func printHeader(h http.Header) {
	fmt.Println("header:")
	for k, v := range h {
		fmt.Printf("  %s:%s\n", k, strings.Join(v, ","))
	}
}

func presignV4(region, ak, sk, method, endpoint, path, hash string, exp time.Duration) error {
	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return err
	}
	req.URL.Path = path

	cred := credentials.NewStaticCredentialsProvider(ak, sk, "")
	sign := v4.NewSigner()

	query := req.URL.Query()
	query.Set("X-Amz-Expires", strconv.FormatInt(int64(exp.Seconds()), 10))
	req.URL.RawQuery = query.Encode()

	u, h, e := sign.PresignHTTP(context.Background(), cred.Value, req, hash, "s3", region, time.Now())
	if e != nil {
		return fmt.Errorf("presign error %w", e)
	}
	printHeader(h)
	fmt.Println("url:\n ", u)
	return nil
}

func signV4(region, ak, sk, method, endpoint, path, hash string, header []string, t time.Time) error {
	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return err
	}
	req.URL.Path = path

	q := req.URL.Query()
	q.Set("x-id", "GetObject")
	req.URL.RawQuery = q.Encode()

	sign := v4.NewSigner()

	for _, v := range header {
		i := strings.Index(v, ":")
		if i < 1 || i >= len(v)-1 {
			return fmt.Errorf("invalid header: %s", v)
		}
		if strings.ToLower(v[:i]) == "x-amz-content-sha256" {
			hash = v[i+1:]
		}
		req.Header.Add(v[:i], v[i+1:])
	}
	req.Header.Add("host", req.URL.Host)

	cred := credentials.NewStaticCredentialsProvider(ak, sk, "")
	e := sign.SignHTTP(context.Background(), cred.Value, req, hash, "s3", "", t)
	if e != nil {
		return fmt.Errorf("sign error %w", e)
	}
	printHeader(req.Header)
	fmt.Println("url:\n ", req.URL.String())
	return nil
}

func main() {
	var endpoint, region, ak, sk string
	var header []string
	var presign bool
	var presignExp time.Duration
	var rootCmd = &cobra.Command{
		Use:   "signer </bucket/key>",
		Short: "signer client tool",
		Long: `S3 sign tool usage:
* presign a GET request
  signer --presign /bucket/key
* sign a GET request
  signer /bucket/key
`,
		Version: version,
		//Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(cmd.Flag("method").Value.String())
			switch method {
			case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPost, http.MethodDelete:
				break
			default:
				return fmt.Errorf("invalid http method: %s", method)
			}
			presign := cmd.Flag("presign").Changed
			hash := cmd.Flag("hash").Value.String()
			t, err := time.Parse("20060102T150405Z", cmd.Flag("time").Value.String())
			if err != nil {
				return fmt.Errorf("invalid time %s %w", cmd.Flag("time").Value.String(), err)
			}

			//path := args[0]
			path := "/open/h1"
			if presign {
				err = presignV4(region, ak, sk, method, endpoint, path, hash, presignExp)
			} else {
				err = signV4(region, ak, sk, method, endpoint, path, hash, header, t)
			}

			if err != nil {
				return err
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "http://192.168.55.2:9000", "S3 endpoint")
	rootCmd.PersistentFlags().StringVarP(&region, "region", "R", "default", "S3 region")
	rootCmd.PersistentFlags().StringVarP(&ak, "ak", "", "object_user1", "access key")
	rootCmd.PersistentFlags().StringVarP(&sk, "sk", "", "ChangeMeChangeMeChangeMeChangeMeChangeMe", "secret key")
	rootCmd.PersistentFlags().BoolVarP(&presign, "presign", "", false, "presign request")
	rootCmd.PersistentFlags().DurationVarP(&presignExp, "expire", "", 12*time.Hour, "presign URL expiration")
	rootCmd.PersistentFlags().StringArrayVarP(&header, "header", "H", []string{"x-amz-content-sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "accept-encoding:identity", "amz-sdk-invocation-id:69ee79c7-a395-45b2-ad71-a7dc6220ec97", "amz-sdk-request:attempt=1; max=3"}, "http headers to sign(-H'x-amz-date:20191119T191919Z')")
	rootCmd.PersistentFlags().StringP("method", "X", http.MethodGet, "http request method")
	rootCmd.PersistentFlags().StringP("hash", "", "UNSIGNED-PAYLOAD", "body checksum")
	//rootCmd.PersistentFlags().StringP("time", "t", time.Now().UTC().Format("20060102T150405Z"), "signing UTC time")
	rootCmd.PersistentFlags().StringP("time", "t", "20201115T144915Z", "signing UTC time")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

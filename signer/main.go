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
	for k, v := range h {
		fmt.Printf("%s:%s\n", k, strings.Join(v, ","))
	}
}

func signV4(region, ak, sk, method, endpoint, path, hash string, presign bool, exp int64) error {
	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return err
	}
	req.URL.Path = path

	v4.NewSigner()

	cred := credentials.NewStaticCredentialsProvider(ak, sk, "")
	sign := v4.NewSigner()

	if presign {
		query := req.URL.Query()
		query.Set("X-Amz-Expires", strconv.FormatInt(exp, 10))
		req.URL.RawQuery = query.Encode()
		u, h, e := sign.PresignHTTP(context.Background(), cred.Value, req, hash, "s3", region, time.Now())
		if e != nil {
			return fmt.Errorf("presign error %w", e)
		}
		printHeader(h)
		fmt.Println(u)
	} else {
		e := sign.SignHTTP(context.Background(), cred.Value, req, "", "s3", region, time.Now())
		if e != nil {
			return fmt.Errorf("sign error %w", e)
		}
		printHeader(req.Header)
		fmt.Println(req.URL.String())
	}
	return nil
}

func main() {
	var endpoint, region, ak, sk string
	var headers []string
	var presign bool
	var rootCmd = &cobra.Command{
		Use:   "sign </bucket/key>",
		Short: "sign client tool",
		Long: `S3 sign tool usage:
* presign a GET request
* sign a GET request
`,
		Version: version,
		Args:    cobra.ExactArgs(1),
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
			err := signV4(region, ak, sk, method, endpoint, args[0], hash, presign, 84600)
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
	rootCmd.PersistentFlags().StringArrayVarP(&headers, "header", "H", nil, "http headers to sign(-H'content-type:application/json' -H'x-amz-a:b')")
	rootCmd.PersistentFlags().StringP("method", "X", http.MethodGet, "http request method")
	rootCmd.PersistentFlags().StringP("hash", "", "UNSIGNED-PAYLOAD", "body checksum")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

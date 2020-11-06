package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
)

var (
	version string = "unknown"
	client  *s3.S3
)

func newsignent(endpoint, region, ak, sk string, pathStyle bool) (*s3.S3, error) {
	fmt.Println("ak: ", ak)
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(ak, sk, ""),
		Endpoint:         &endpoint,
		Region:           &region,
		S3ForcePathStyle: &pathStyle,
	})
	if err != nil {
		return nil, err
	}

	// Create a new instance of the service's client with a Session.
	// Optional aws.Config values can also be provided as variadic arguments
	// to the New function. This option allows you to provide service
	// specific configuration.
	svc := s3.New(sess)

	return svc, nil
}

func sign(name, method, path string, presign bool) error {
	op := &request.Operation{
		Name:       name,
		HTTPMethod: method,
		HTTPPath:   path,
	}

	req := client.NewRequest(op, nil, nil)

	if presign {
		s, h, e := req.PresignRequest(86400 * time.Second)
		if e != nil {
			return e
		}
		fmt.Println(h)
		fmt.Println(s)
	} else {
		e := req.Sign()
		if e != nil {
			return e
		}
		fmt.Println(req.HTTPRequest.Header)
		fmt.Println(req.HTTPRequest.URL.String())
	}
	return nil
}

func main() {
	var endpoint, region, ak, sk string
	var debug, verbose, pathStyle, presign bool
	var err error
	var rootCmd = &cobra.Command{
		Use:   "sign",
		Short: "sign client tool",
		Long: `S3 command-line tool usage:
Endpoint EnvVar:
	S3_ENDPOINT=http://host:port (only read if flag -e is not set)

Credential EnvVar:
	AWS_ACCESS_KEY_ID=AK      (only read if flag -p is not set or --ak is not set)
	AWS_ACCESS_KEY=AK         (only read if AWS_ACCESS_KEY_ID is not set)
	AWS_SECRET_ACCESS_KEY=SK  (only read if flag -p is not set or --sk is not set)
	AWS_SECRET_KEY=SK         (only read if AWS_SECRET_ACCESS_KEY is not set)`,
		Version: version,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err = newsignent(endpoint, region, ak, sk, pathStyle)
			if err != nil {
				return err
			}
			method := strings.ToUpper(cmd.Flag("method").Value.String())
			switch method {
			case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPost, http.MethodDelete:
				break
			default:
				return fmt.Errorf("invalid http method: %s", method)
			}
			presign := cmd.Flag("presign").Changed
			var s string
			var err error
			//contentType := cmd.Flag("content-type").Value.String()
			err = sign("GetObject", method, args[0], presign)
			if err != nil {
				return err
			}
			fmt.Println(s)
			return nil
		},
	}
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "", false, "print debug log")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	rootCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "http://192.168.55.2:9000", "S3 endpoint")
	rootCmd.PersistentFlags().StringVarP(&region, "region", "R", "default", "S3 region")
	rootCmd.PersistentFlags().StringVarP(&ak, "ak", "", "object_user1", "access key")
	rootCmd.PersistentFlags().StringVarP(&sk, "sk", "", "ChangeMeChangeMeChangeMeChangeMeChangeMe", "secret key")
	rootCmd.PersistentFlags().BoolVarP(&pathStyle, "path-style", "", true, "use path style(not use virtual-host)")
	rootCmd.PersistentFlags().BoolVarP(&presign, "presign", "", false, "presign request")

	rootCmd.PersistentFlags().StringSliceP("header", "H", []string{"content-type:text"}, "http headers")
	rootCmd.PersistentFlags().StringP("method", "X", http.MethodGet, "http request method")
	rootCmd.PersistentFlags().StringP("content-type", "T", "", "http request content-type")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

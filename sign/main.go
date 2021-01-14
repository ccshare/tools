package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	version string = "unknown"
)

func newS3Client(endpoint, region, ak, sk string) *s3.Client {
	// Load the SDK's configuration from environment and shared config, and
	// create the client with this.
	cfg, err := config.LoadDefaultConfig(
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     ak,
				SecretAccessKey: sk,
				Source:          "provider",
			},
		}),
	)

	if err != nil {
		log.Fatalf("failed to load SDK configuration, %v", err)
	}

	cfg.EndpointResolver = aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: endpoint,
		}, nil

	})
	cfg.Region = region

	opts := s3.Options{
		Region:           region,
		Credentials:      cfg.Credentials,
		EndpointResolver: s3.WithEndpointResolver(cfg.EndpointResolver, s3.NewDefaultEndpointResolver()),
		UsePathStyle:     true,
	}
	return s3.New(opts)
}

func catObject(region, ak, sk, method, endpoint, bucket, key string) error {
	client := newS3Client(endpoint, region, ak, sk)
	so, err := client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, so.Body)
	return nil
}

func main() {
	region := "default"
	ak := "object_user1"
	sk := "ChangeMeChangeMeChangeMeChangeMeChangeMe"
	method := "GET"
	endpoint := "http://192.168.55.2:9000"
	bucket := "open"
	key := "h1"

	err := catObject(region, ak, sk, method, endpoint, bucket, key)
	if err != nil {
		fmt.Println("error: ", err)
		return
	}
	return

}

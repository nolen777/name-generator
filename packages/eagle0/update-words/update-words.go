package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type httpInfo struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type Event struct {
	Http httpInfo `json:"http"`
}

var bucketName = "eagle0-config"

func UpdateWords(ctx context.Context, event Event) {
	endpoint := "sfo3.digitaloceanspaces.com"
	region := "sfo3"

	accessKeyId := os.Getenv("DIGITALOCEAN_ACCESS_KEY_ID")
	secretKey := os.Getenv("DIGITALOCEAN_SECRET_KEY")

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKeyId, secretKey, ""),
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(false),
		Region:           aws.String(region),
	}
	sess, err := session.NewSession(s3Config)
	if err != nil {
		panic(err)
	}
	s3Client := s3.New(sess)

	t := time.Now()
	content := t.Format("2006-01-02 15:04:05")

	putObjInput := s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("sample.txt"),
		Body:   aws.ReadSeekCloser(strings.NewReader(content)),
	}

	out, err := s3Client.PutObject(&putObjInput)
	if err != nil {
		fmt.Println("Error uploading file:", err)
	} else {
		fmt.Println("File uploaded successfully:", out)
	}
}

package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
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

	httpRead, err := http.Get("https://docs.google.com/spreadsheets/d/1DHEsiv4cY4gE6AX3sVH82K__mpBD1aznIYCQwQxA_F0/export?gid=0&format=tsv")
	if err != nil {
		fmt.Println("Error fetching file:", err)
		return
	}

	content, err := io.ReadAll(httpRead.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	toWrite := condensedTsv(string(content))

	putObjInput := s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("names.tsv"),
		Body:   strings.NewReader(toWrite),
	}

	out, err := s3Client.PutObject(&putObjInput)
	if err != nil {
		fmt.Println("Error uploading file:", err)
	} else {
		fmt.Println("File uploaded successfully:", out)
	}
}

func condensedTsv(tsv string) string {
	lines := strings.Split(tsv, "\r\n")
	headers := strings.Split(lines[0], "\t")
	namesMap := make(map[string]map[string]bool)
	for _, h := range headers {
		namesMap[h] = make(map[string]bool)
	}

	for _, line := range lines[1:] {
		if len(line) == 0 {
			continue
		}
		values := strings.Split(line, "\t")
		for i, h := range headers {
			namesMap[h][strings.TrimSpace(values[i])] = true
		}
	}

	sort.Slice(headers, func(i, j int) bool {
		return len(namesMap[headers[i]]) > len(namesMap[headers[j]])
	})

	sortedNamesMap := make(map[string][]string)
	for _, h := range headers {
		sortedNamesMap[h] = make([]string, 0, len(namesMap[h]))
		for name := range namesMap[h] {
			if name == "" {
				continue
			}
			sortedNamesMap[h] = append(sortedNamesMap[h], name)
		}
		sort.Strings(sortedNamesMap[h])
	}

	var condensedLines []string
	condensedLines = append(condensedLines, strings.Join(headers, "\t"))
	for i := 0; ; i++ {
		lineWords := []string{}
		for _, h := range headers {
			if i < len(sortedNamesMap[h]) {
				lineWords = append(lineWords, sortedNamesMap[h][i])
			}
		}
		if len(lineWords) == 0 {
			break
		}
		condensedLines = append(condensedLines, strings.Join(lineWords, "\t"))
	}

	return strings.Join(condensedLines, "\r\n")
}

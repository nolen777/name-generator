package spaces_fetcher

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"os"
)

var bucketName = "eagle0-config"
var sharedClient *s3.S3 = createS3Client()

func createS3Client() *s3.S3 {
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

	return s3Client
}

func GetFile(path string) ([]byte, error) {
	getObjInput := s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(path),
	}

	out, err := sharedClient.GetObject(&getObjInput)
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()

	return io.ReadAll(out.Body)
}

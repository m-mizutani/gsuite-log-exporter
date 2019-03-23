package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/pkg/errors"
)

func getSecretValues(secretArn string, values interface{}) error {
	// sample: arn:aws:secretsmanager:ap-northeast-1:1234567890:secret:mytest
	arn := strings.Split(secretArn, ":")
	if len(arn) != 7 {
		return errors.New(fmt.Sprintf("Invalid SecretsManager ARN format: %s", secretArn))
	}
	region := arn[3]

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	mgr := secretsmanager.New(ssn)

	result, err := mgr.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretArn),
	})

	if err != nil {
		return errors.Wrap(err, "Fail to retrieve secret values")
	}

	err = json.Unmarshal([]byte(*result.SecretString), values)
	if err != nil {
		return errors.Wrap(err, "Fail to parse secret values as JSON")
	}

	return nil
}

type s3Uploader struct {
	svc     *s3.S3
	args    Arguments
	wg      sync.WaitGroup
	chQueue chan *queue
}

func newS3Uploader(args Arguments) *s3Uploader {
	threadNum := 16

	uploader := new(s3Uploader)

	uploader.svc = s3.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String(args.S3Region),
	})))
	uploader.args = args

	uploader.chQueue = make(chan *queue, threadNum*2)

	for i := 0; i < threadNum; i++ {
		uploader.wg.Add(1)
		go putWorker(uploader.chQueue, args, uploader.svc, &uploader.wg)
	}

	return uploader
}

func putWorker(chQueue chan *queue, args Arguments, svc *s3.S3, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		q, ok := <-chQueue
		if !ok {
			return
		}

		s3Key := strings.Join([]string{
			args.S3Prefix, q.app, q.timestamp.Format("/2006/01/02/15/"),
			q.timestamp.Format("20060102_150405_"), q.key, ".json.gz"}, "")

		_, err := svc.HeadObject(&s3.HeadObjectInput{
			Bucket: &args.S3Bucket,
			Key:    &s3Key,
		})

		exists := true
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case s3.ErrCodeNoSuchKey:
					exists = false
				case "NotFound":
					exists = false
				default:
					logger.WithError(err).Fatalf("HeadObject error: %s", aerr.Error())
				}
			} else {
				logger.WithError(err).Fatalf("HeadObject error")
			}
		}

		if !exists {
			var buf bytes.Buffer
			zw := gzip.NewWriter(&buf)
			zw.Write(q.data)
			zw.Close()

			_, err = svc.PutObject(&s3.PutObjectInput{
				Body:   bytes.NewReader(buf.Bytes()),
				Bucket: &args.S3Bucket,
				Key:    &s3Key,
			})

			if err != nil {
				logger.WithError(err).Fatalf("Fail to put log object: %s", s3Key)
			}
		}
	}
}

func (x *s3Uploader) putLogObject(q *queue) {
	x.chQueue <- q
}

func (x *s3Uploader) wait() {
	close(x.chQueue)
	x.wg.Wait()
}

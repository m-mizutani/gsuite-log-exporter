package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

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

func putLogObject(svc *s3.S3, q *queue, args Arguments) (bool, error) {
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
				return false, errors.Wrapf(err, "HeadObject error: %s", aerr.Error())
			}
		} else {
			return false, err
		}
	}

	if !exists {
		_, err := svc.PutObject(&s3.PutObjectInput{
			Body:   bytes.NewReader(q.data),
			Bucket: &args.S3Bucket,
			Key:    &s3Key,
		})

		if err != nil {
			return false, errors.Wrapf(err, "Fail to put log object: %s", s3Key)
		}

		return true, nil
		fmt.Println("upload")
	}

	return false, nil
}

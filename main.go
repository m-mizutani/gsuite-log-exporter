package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var loggerBase = logrus.New()
var logger = loggerBase.WithFields(logrus.Fields{})

type Arguments struct {
	SecretArn string `json:"secret_arn"`
	S3Region  string `json:"s3_region"`
	S3Bucket  string `json:"s3_bucket"`
	S3Prefix  string `json:"s3_prefix"`
}

type Response struct {
	LogCount int
}

type secretValues struct {
	GSuiteClient string `json:"gsuite_client"`
	GSuiteToken  string `json:"gsuite_token"`
}

func Handler(args Arguments) (*Response, error) {
	var resp Response
	var secrets secretValues

	if err := getSecretValues(args.SecretArn, &secrets); err != nil {
		return nil, errors.Wrap(err, "Fail to fetch values from SecretsManager")
	}

	client, err := setupGoogleClient([]byte(secrets.GSuiteClient), []byte(secrets.GSuiteToken))
	if err != nil {
		return nil, err
	}

	uploader := newS3Uploader(args)

	for q := range exportLogs(client) {
		if q.err != nil {
			return nil, q.err
		}

		resp.LogCount++
		uploader.putLogObject(q)
	}

	return &resp, nil
}

func handleRequest(ctx context.Context, event events.S3Event) error {
	lc, _ := lambdacontext.FromContext(ctx)
	logger = loggerBase.WithField("request_id", lc.AwsRequestID)

	logger.WithField("event", event).Info("Start")

	args := Arguments{
		SecretArn: os.Getenv("SECRET_ARN"),
		S3Region:  os.Getenv("S3_REGION"),
		S3Bucket:  os.Getenv("S3_BUCKET"),
		S3Prefix:  os.Getenv("S3_PREFIX"),
	}

	resp, err := Handler(args)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"args":  args,
			"resp":  resp,
			"error": err,
		}).Error("Fail to export G Suite log")
		return err
	}

	logger.WithField("resp", resp).Info("Exit")
	return nil
}

func main() {
	loggerBase.SetLevel(logrus.InfoLevel)
	loggerBase.SetFormatter(&logrus.JSONFormatter{})
	lambda.Start(handleRequest)
}

package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/sirupsen/logrus"
)

var loggerBase = logrus.New()
var logger = loggerBase.WithFields(logrus.Fields{})

func handleRequest(ctx context.Context, event events.S3Event) error {
	lc, _ := lambdacontext.FromContext(ctx)
	logger = loggerBase.WithField("request_id", lc.AwsRequestID)

	logger.WithField("event", event).Info("Start")

	return nil
}

func main() {
	loggerBase.SetLevel(logrus.InfoLevel)
	loggerBase.SetFormatter(&logrus.JSONFormatter{})
	lambda.Start(handleRequest)
}

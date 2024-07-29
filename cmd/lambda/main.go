package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
)

type Event struct {
	S3Bucket string `json:"s3Bucket"`
	S3Key    string `json:"s3Key"`
}

type FaceAlreadyExistsError struct{}

func (e *FaceAlreadyExistsError) Error() string {
	return "Face in the picture is already in the system."
}

func handler(ctx context.Context, event Event) (string, error) {
	log.Println("Reading input from event:", event)

	srcBucket := event.S3Bucket
	srcKey, err := url.QueryUnescape(strings.Replace(event.S3Key, "+", " ", -1))
	if err != nil {
		return "", fmt.Errorf("failed to decode S3 key: %v", err)
	}

	sess := session.Must(session.NewSession())
	rekognitionClient := rekognition.New(sess)

	params := &rekognition.SearchFacesByImageInput{
		CollectionId:       aws.String(os.Getenv("REKOGNITION_COLLECTION_ID")),
		Image:              &rekognition.Image{S3Object: &rekognition.S3Object{Bucket: aws.String(srcBucket), Name: aws.String(srcKey)}},
		FaceMatchThreshold: aws.Float64(70.0),
		MaxFaces:           aws.Int64(3),
	}

	result, err := rekognitionClient.SearchFacesByImage(params)
	if err != nil {
		return "", fmt.Errorf("failed to search faces by image: %v", err)
	}

	log.Println("Search results:", result)

	if len(result.FaceMatches) > 0 {
		return "", &FaceAlreadyExistsError{}
	}

	return "No matching faces found.", nil
}

func main() {
	lambda.Start(handler)
}

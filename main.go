package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ehpalumbo/go-diff/api"
	"github.com/ehpalumbo/go-diff/domain"
	"github.com/ehpalumbo/go-diff/repository"
	"github.com/ehpalumbo/go-diff/service"
)

func main() {
	repo := repository.NewS3DiffRepository(getS3Client(), os.Getenv("AWS_BUCKET_NAME"))
	app := RunApplication(repo)
	lambda.Start(func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return app.Handle(req), nil
	})
}

// RunApplication is the application entrypoint that provides the lambda handler.
// Requires a DiffRepository implementation.
func RunApplication(repo service.DiffRepository) api.Application {
	diff := domain.NewDifferImpl()
	svc := service.NewDiffService(diff, repo)
	return api.NewApplication(svc)
}

func getS3Client() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Cannot load AWS configuration.\n", err)
	}
	return s3.NewFromConfig(cfg)
}

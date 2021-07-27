package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/ehpalumbo/go-diff/api"
	"github.com/ehpalumbo/go-diff/domain"
	"github.com/ehpalumbo/go-diff/repository"
	"github.com/ehpalumbo/go-diff/service"
)

func main() {
	repo := repository.NewS3DiffRepository(getS3Client(), os.Getenv("AWS_BUCKET_NAME"))
	handler := initLambdaHandler(repo)
	lambda.Start(handler)
}

type LambdaHandler func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// initLambdaHandler is the application entrypoint that provides the lambda handler.
// Requires a DiffRepository implementation.
func initLambdaHandler(repo service.DiffRepository) LambdaHandler {
	diff := domain.NewDifferImpl()
	svc := service.NewDiffService(diff, repo)
	app := api.NewApplication(svc)
	adapter := ginadapter.New(app.GetRouter())
	return adapter.Proxy
}

func getS3Client() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Cannot load AWS configuration.\n", err)
	}
	return s3.NewFromConfig(cfg)
}

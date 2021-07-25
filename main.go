package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ehpalumbo/go-diff/api"
	"github.com/ehpalumbo/go-diff/domain"
	"github.com/ehpalumbo/go-diff/repository"
	"github.com/ehpalumbo/go-diff/service"
)

func main() {
	repo := repository.NewS3DiffRepository(getS3Client(), os.Getenv("AWS_BUCKET_NAME"))
	shutdown := RunApplication(":8080", repo)
	// graceful shutdown is achieved by listening to system signals
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	// try to shutdown server in 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	shutdown(ctx)
}

// RunApplication is the application entrypoint.
// Requires a DiffRepository implementation.
// It returns a function to shutdown the started service.
func RunApplication(addr string, repo service.DiffRepository) func(context.Context) {
	diff := domain.NewDifferImpl()
	svc := service.NewDiffService(diff, repo)
	app := api.NewApplication(svc)
	srv := &http.Server{Addr: addr, Handler: app.GetRouter()}
	go run(srv) // run async: avoid blocking this coroutine
	return shutdownFunc(srv)
}

func run(srv *http.Server) {
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server stopped listening.\n", err)
	}
}

func shutdownFunc(srv *http.Server) func(context.Context) {
	return func(ctx context.Context) {
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server forced to shutdown.\n", err)
		}
	}
}

func getS3Client() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Cannot load AWS configuration.\n", err)
	}
	return s3.NewFromConfig(cfg)
}

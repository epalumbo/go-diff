package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ehpalumbo/go-diff/api"
	"github.com/ehpalumbo/go-diff/domain"
	"github.com/ehpalumbo/go-diff/repository"
	"github.com/ehpalumbo/go-diff/service"
	"github.com/go-redis/redis/v8"
)

func main() {
	shutdown := RunApplication(":8080", os.Getenv("REDIS_URL"))
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
// It takes the Redis URL as input but it defaults to localhost if not defined.
// It returns a function to shutdown the started service.
func RunApplication(addr, dbURL string) func(context.Context) {
	repo := repository.NewRedisRepository(getRedisClientOptions(dbURL))
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

func getRedisClientOptions(dbURL string) *redis.Options {
	if len(dbURL) == 0 {
		dbURL = "redis://localhost:6379/"
	}
	opt, err := redis.ParseURL(dbURL)
	if err != nil {
		panic(err)
	}
	return opt
}

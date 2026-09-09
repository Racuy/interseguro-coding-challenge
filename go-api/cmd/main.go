// go-api: rotates a matrix, computes its QR, sends it to node-api, returns the stats
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"go-api/internal/application"
	"go-api/internal/controller/auth"
	"go-api/internal/controller/gateway"
	httpController "go-api/internal/controller/http"
	"go-api/internal/controller/middleware"
)

// grace period before force-closing in-flight requests on shutdown
const shutdownGracePeriod = 10 * time.Second

func main() {
	app := fiber.New(fiber.Config{
		AppName: "go-api",
	})

	app.Use(recover.New())
	app.Use(logger.New())

	nodeAPIURL := requireEnv("NODE_API_URL")
	jwtSecret := requireEnv("JWT_SECRET")

	tokenIssuer := auth.NewJWTIssuer(jwtSecret)
	authUsecase := application.NewAuthUsecase(tokenIssuer)
	authHandler := httpController.NewAuthHandler(authUsecase)

	statsGateway := gateway.NewNodeStatsGateway(nodeAPIURL)
	matrixUsecase := application.NewMatrixUsecase(statsGateway)
	matrixHandler := httpController.NewMatrixHandler(matrixUsecase)

	api := app.Group("/api/v1")
	httpController.RegisterAuthRoutes(api, authHandler)

	protected := api.Group("", middleware.JWTProtected(jwtSecret))
	httpController.RegisterRoutes(protected, matrixHandler)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// run in background, wait for either a startup error or a stop signal
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- app.Listen(":8080")
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("server failed to start: %v", err)

	case sig := <-quit:
		// drain in-flight requests before exiting
		log.Printf("received %s, shutting down...", sig)

		ctx, cancel := context.WithTimeout(context.Background(), shutdownGracePeriod)
		defer cancel()

		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Printf("forced shutdown after %s: %v", shutdownGracePeriod, err)
		}
	}
}

// reads an env var, exits if it's not set
func requireEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s environment variable is required", name)
	}
	return value
}

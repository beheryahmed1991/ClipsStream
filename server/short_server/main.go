package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/beheryahmed1991/ClipsStream/server/short_server/controllers"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
)

type HelloOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

// main initializes logging and the HTTP server, registers API routes under /api (including GET /hello which returns {"message":"hello"}), and starts the server on :8080.
// If the server fails to start, main logs the error and exits with status 1.
func main() {
	gin.SetMode(gin.ReleaseMode)
	// r:=gin.Default()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	apiGroup := r.Group("/api")

	config := huma.DefaultConfig("My API", "1.0.0")
	config.Servers = []*huma.Server{
		{URL: "/api"},
	}

	api := humagin.NewWithGroup(r, apiGroup, config)

	huma.Get(api, "/hello", func(ctx context.Context, in *struct{}) (*HelloOutput, error) {
		out := &HelloOutput{}
		out.Body.Message = "hello"
		return out, nil
	})
	controllers.RegisterMovRoutes(api)
	slog.Info("server starting", "addr", ":8080")
	if err := r.Run(":8080"); err != nil {
		slog.Error("server failed to start", "err", err)
		os.Exit(1)
	}
}
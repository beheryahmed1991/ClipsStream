package main

import (
	"context"
	"log"
	"os"

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

func main() {
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
	if err := r.Run(":8080"); err != nil {
		log.Printf("server failed to start: %v", err)
		os.Exit(1)
	}
}

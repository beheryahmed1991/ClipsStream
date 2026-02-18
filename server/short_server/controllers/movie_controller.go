package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/beheryahmed1991/ClipsStream/server/short_server/database"
	model "github.com/beheryahmed1991/ClipsStream/server/short_server/models"
	"github.com/danielgtaylor/huma/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type (
	// what datas come out from the endpoint
	GetMoviesOutput struct {
		Body []model.Movie `json:"body"`
	}

	GetMovieOutput struct {
		Body model.Movie `json:"body"`
	}

	// what data put in
	GetMovieInput struct {
		ID string `path:"id"`
	}
)

func RegisterMovRoutes(api huma.API) {
	huma.Get(api, "/movies", GetMovies)
	huma.Get(api, "/movies/{id}", GetMovie)
}

func GetMovies(ctx context.Context, in *struct{}) (*GetMoviesOutput, error) {
	col, err := database.OpenCollection("movies")
	if err != nil {
		return nil, fmt.Errorf("open movies collection: %w", err)
	}

	qctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cursor, err := col.Find(qctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("find moives: %w", err)
	}
	defer cursor.Close(qctx)

	var movies []model.Movie

	if err := cursor.All(qctx, &movies); err != nil {
		return nil, fmt.Errorf("decode movies: %w", err)
	}

	return &GetMoviesOutput{Body: movies}, nil
}

func GetMovie(ctx context.Context, in *GetMovieInput) (*GetMovieOutput, error) {
	col, err := database.OpenCollection("movies")
	if err != nil {
		return nil, fmt.Errorf("open collection movies %w", err)
	}
	objID, err := bson.ObjectIDFromHex(in.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid movie id:%w", err)
	}
	qctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var movie model.Movie
	if err := col.FindOne(qctx, bson.M{"_id": objID}).Decode(&movie); err != nil {
		return nil, fmt.Errorf("find movie: %w", err)
	}

	return &GetMovieOutput{Body: movie}, nil
}

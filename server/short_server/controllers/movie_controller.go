package controllers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/beheryahmed1991/ClipsStream/server/short_server/database"
	model "github.com/beheryahmed1991/ClipsStream/server/short_server/models"
	"github.com/danielgtaylor/huma/v2"

	//"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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

var (
	validate = validator.New()
)

func RegisterMovRoutes(api huma.API) {
	huma.Get(api, "/movies", GetMovies)
	//huma.Get(api, "/movies/{id}", GetMovie)
	huma.Register(api, huma.Operation{
		OperationID: "get-movie",
		Method:      "GET",
		Path:        "/movies/{id}",
		Summary:     "Get one movie by ID",
		Errors:      []int{400, 404},
	}, GetMovie)
}

func getMovieCol() (*mongo.Collection, error) {
	return database.OpenCollection("movies")
}

func GetMovies(ctx context.Context, in *struct{}) (*GetMoviesOutput, error) {
	col, err := getMovieCol()
	if err != nil {
		slog.Error("open movies collection failed", "op", "GetMovies", "err", err)
		return nil, fmt.Errorf("open movies collection: %w", err)
	}

	qctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cursor, err := col.Find(qctx, bson.M{})
	if err != nil {
		slog.Error("find movies failed", "op", "GetMovies", "err", err)
		return nil, fmt.Errorf("find movies: %w", err)
	}
	defer cursor.Close(qctx)

	movies := make([]model.Movie, 0)

	if err := cursor.All(qctx, &movies); err != nil {
		slog.Error("decode movies failed", "op", "GetMovies", "err", err)
		return nil, fmt.Errorf("decode movies: %w", err)
	}

	return &GetMoviesOutput{Body: movies}, nil
}

func GetMovie(ctx context.Context, in *GetMovieInput) (*GetMovieOutput, error) {
	col, err := getMovieCol()
	if err != nil {
		slog.Error("open movies collection failed", "op", "GetMovie", "movie_id", in.ID, "err", err)
		return nil, fmt.Errorf("open collection movies %w", err)
	}
	objID, err := bson.ObjectIDFromHex(in.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid movie ID")
	}
	qctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var movie model.Movie
	if err := col.FindOne(qctx, bson.M{"_id": objID}).Decode(&movie); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, huma.Error404NotFound("movie not found")
		}
		slog.Error("find movie failed", "op", "GetMovie", "movie_id", in.ID, "err", err)
		return nil, fmt.Errorf("find movie: %w", err)
	}

	return &GetMovieOutput{Body: movie}, nil
}

// function to add moive
/* func AddMovie() gin.HandlersChain {

	return nil
} */

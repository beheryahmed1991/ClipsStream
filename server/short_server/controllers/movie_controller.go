package controllers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/beheryahmed1991/ClipsStream/server/short_server/database"
	model "github.com/beheryahmed1991/ClipsStream/server/short_server/models"
	"github.com/danielgtaylor/huma/v2"
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

	// for post
	AddMovieInput struct {
		Body model.Movie
	}
	AddMovieOutput struct {
		Body model.Movie `json:"body"`
	}
)

var (
	validate = validator.New()
)

// RegisterMovRoutes registers HTTP routes for movie-related operations on the provided Huma API.
// It registers GET /movies, GET /movies/{id} (operation ID "get-movie"), and POST /addmovies (operation ID "add-movie")
// which uses 201 Created as the default response status.
func RegisterMovRoutes(api huma.API) {
	huma.Get(api, "/movies", GetMovies)
	//huma.Get(api, "/movies/{id}", GetMovie)
	huma.Register(api, huma.Operation{
		OperationID: "get-movie",
		Method:      "GET",
		Path:        "/movies/{id}",
		Summary:     "Get one movie by ID",
		Errors:      []int{400, 404, 500},
	}, GetMovie)
	huma.Register(api, huma.Operation{
		OperationID:   "add-movie",
		Method:        "POST",
		Path:          "/addmovies",
		Summary:       "Add one movie",
		DefaultStatus: http.StatusCreated,
		Errors:        []int{400, 500},
	}, AddMovie)
}

// getMovieCol returns the MongoDB collection used for storing movies.
// It returns an error if the collection cannot be opened.
func getMovieCol() (*mongo.Collection, error) {
	return database.OpenCollection("movies")
}

// GetMovies retrieves all movies from the database.
// It returns a GetMoviesOutput whose Body contains the retrieved movies, or an error if the operation fails.
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

// GetMovie retrieves the movie identified by the provided hex ID from the movies collection.
// 
// It returns the movie in a GetMovieOutput on success.
// If the provided ID is not a valid hex ObjectID, it returns a 400 Bad Request error.
// If no movie with the given ID exists, it returns a 404 Not Found error.
// Other failures (collection open or database errors) are returned as wrapped errors.
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

// AddMovie validates the provided movie, assigns it a new BSON ObjectID, inserts it into the movies collection, and returns the created movie.
// If validation fails it returns a 400 Bad Request error with field-specific details. It returns an error if opening the collection or inserting the document fails.
func AddMovie(ctx context.Context, in *AddMovieInput) (*AddMovieOutput, error) {
	if err := validate.Struct(in.Body); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			details := make([]error, len(ve))
			for i, fe := range ve {
				details[i] = fmt.Errorf("field '%s' failed '%s'", fe.Field(), fe.Tag())
			}
			return nil, huma.Error400BadRequest("validation failed", details...)
		}
		return nil, huma.Error400BadRequest("validation failed")
	}
	col, err := getMovieCol()
	if err != nil {
		slog.Error("open movies collection failed", "op", "AddMovie", "err", err)
		return nil, fmt.Errorf("open movies collection: %w", err)
	}
	qctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	movie := in.Body
	movie.ID = bson.NewObjectID()

	if _, err := col.InsertOne(qctx, movie); err != nil {
		slog.Error("insert movie failed", "op", "AddMovie", "err", err)
		return nil, fmt.Errorf("insert movie: %w", err)
	}
	return &AddMovieOutput{
		Body: movie,
	}, nil

}
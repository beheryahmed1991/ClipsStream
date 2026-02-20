package controllers

import (
	"context"
	"testing"

	model "github.com/beheryahmed1991/ClipsStream/server/short_server/models"
)

func TestContronller(t *testing.T) {
	t.Run("test controller get all moives", func(t *testing.T) {
		out, err := GetMovie(context.Background(), &GetMovieInput{ID: "bad-id"})
		if err == nil {
			t.Fatal("expected error for invalid id")
		}
		if out != nil {
			t.Fatal("expected nil output")
		}

	})
}

func TestAddMovie(t *testing.T) {
	t.Run("returns error for invalid payload", func(t *testing.T) {
		out, err := AddMovie(context.Background(), &AddMovieInput{
			Body: model.Movie{
				Title: "",
			},
		})
		if err == nil {
			t.Fatal("expected error for invalid movie payload")
		}
		if out != nil {
			t.Fatal("expected nil output")
		}
	})
}

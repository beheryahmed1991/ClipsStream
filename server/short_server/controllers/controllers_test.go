package controllers

import (
	"context"
	"testing"

	model "github.com/beheryahmed1991/ClipsStream/server/short_server/models"
	"golang.org/x/crypto/bcrypt"
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

	t.Run("test controller get one user with invalid id", func(t *testing.T) {
		out, err := GetUser(context.Background(), &GetUserInput{ID: "bad-id"})
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

func TestAddUser(t *testing.T) {
	t.Run("returns error for invalid payload", func(t *testing.T) {
		out, err := AddUser(context.Background(), &AddUserInput{
			Body: AddUserRequestBody{
				Email: "invalid",
			},
		})
		if err == nil {
			t.Fatal("expected error for invalid user payload")
		}
		if out != nil {
			t.Fatal("expected nil output")
		}
	})
}

func TestHashPassword(t *testing.T) {
	plain := "secret123"
	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hash == plain {
		t.Fatal("expected password to be hashed")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		t.Fatalf("hash does not match plaintext: %v", err)
	}
}

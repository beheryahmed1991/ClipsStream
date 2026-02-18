package controllers

import (
	"context"
	"testing"
)

func TestContronller(t *testing.T) {
	t.Run("test controller get all moives", func(t *testing.T) {
		out, err := GetMovie(context.Background(), &GetMovieInput{ID: "6991aba260bb00b9d434ac58"})
		if err == nil {
			t.Fatal("expectd error for invaled id")
		}
		if out != nil {
			t.Fatal("expected nil output")
		}

	})
}

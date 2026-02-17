package database

import (
	"errors"
	"sync"
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func resetDBState() {
	clientOnce = sync.Once{}
	client = nil
	clientErr = nil
}

// test for create connection
func TestInstance(t *testing.T) {
	resetDBState()
	origConnect := connectMongo
	origPing := pingMongo
	defer func() {
		connectMongo = origConnect
		pingMongo = origPing
	}()
	t.Setenv("MONGODB_URI", "mongodb://fake")
	connectMongo = func(uri string) (*mongo.Client, error) {
		return nil, errors.New("connect fails")
		//test fail
		//return &mongo.Client{}, nil

	}
	c, err := InstanceDB()
	//test fail
	//if err != nil {

	if err == nil || c != nil {
		t.Fatalf("expectd connect error, got client=%v err=%v", c, err)
	}

}

func TestOpenColl(t *testing.T) {
	resetDBState()
	col, err := OpenCollection("")
	//fail	if err != nil { check of the empty files
	if err == nil {
		t.Fatalf("expectd error for empty collection, got nil. collection= %v", col)
	}
	// fail if col == nil { check if the object was returned on failure !!}
	if col != nil {
		t.Fatalf("expectd nil collection, got %v", col)
	}
}

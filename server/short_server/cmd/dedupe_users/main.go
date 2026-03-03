package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/beheryahmed1991/ClipsStream/server/short_server/database"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type duplicateGroup struct {
	Email string          `bson:"_id"`
	IDs   []bson.ObjectID `bson:"ids"`
	Count int             `bson:"count"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	col, err := database.OpenCollection("users")
	if err != nil {
		log.Fatalf("open users collection: %v", err)
	}

	removed, groups, err := removeDuplicateUsers(ctx, col)
	if err != nil {
		log.Fatalf("remove duplicates: %v", err)
	}

	if err := normalizeEmails(ctx, col); err != nil {
		log.Fatalf("normalize emails: %v", err)
	}

	if err := ensureUniqueEmailIndex(ctx, col); err != nil {
		log.Fatalf("ensure unique email index: %v", err)
	}

	fmt.Printf("Done. Duplicate groups: %d, removed users: %d\n", groups, removed)
}

func removeDuplicateUsers(ctx context.Context, col *mongo.Collection) (int, int, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"email": bson.M{
				"$type": "string",
				"$ne":   "",
			},
		}}},
		{{Key: "$addFields", Value: bson.M{
			"norm_email": bson.M{
				"$toLower": bson.M{
					"$trim": bson.M{"input": "$email"},
				},
			},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$norm_email",
			"ids":   bson.M{"$push": "$_id"},
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$match", Value: bson.M{
			"_id":   bson.M{"$ne": ""},
			"count": bson.M{"$gt": 1},
		}}},
	}

	cur, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0, err
	}
	defer cur.Close(ctx)

	removed := 0
	groups := 0

	for cur.Next(ctx) {
		var g duplicateGroup
		if err := cur.Decode(&g); err != nil {
			return removed, groups, err
		}
		if len(g.IDs) <= 1 {
			continue
		}

		groups++
		deleteIDs := g.IDs[1:]
		res, err := col.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": deleteIDs}})
		if err != nil {
			return removed, groups, err
		}
		removed += int(res.DeletedCount)
	}

	if err := cur.Err(); err != nil {
		return removed, groups, err
	}
	return removed, groups, nil
}

func normalizeEmails(ctx context.Context, col *mongo.Collection) error {
	_, err := col.UpdateMany(
		ctx,
		bson.M{"email": bson.M{"$type": "string"}},
		mongo.Pipeline{
			{{Key: "$set", Value: bson.M{
				"email": bson.M{
					"$toLower": bson.M{
						"$trim": bson.M{"input": "$email"},
					},
				},
			}}},
		},
	)
	return err
}

func ensureUniqueEmailIndex(ctx context.Context, col *mongo.Collection) error {
	model := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetName("users_email_unique").SetUnique(true),
	}
	_, err := col.Indexes().CreateOne(ctx, model)
	return err
}


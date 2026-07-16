package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"zametka/internal/config"
)

const connectTimeout = 10 * time.Second

func Connect(ctx context.Context, cfg config.Config) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	pingCtx, pingCancel := context.WithTimeout(ctx, connectTimeout)
	defer pingCancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	return client, nil
}

func EnsureIndexes(ctx context.Context, client *mongo.Client, dbName string) error {
	db := client.Database(dbName)

	rooms := db.Collection("rooms")
	_, err := rooms.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("rooms index: %w", err)
	}

	notes := db.Collection("notes")
	_, err = notes.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "roomId", Value: 1},
			{Key: "createdAt", Value: -1},
		},
	})
	if err != nil {
		return fmt.Errorf("notes roomId+createdAt index: %w", err)
	}

	_, err = notes.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "roomId", Value: 1},
			{Key: "category", Value: 1},
			{Key: "createdAt", Value: -1},
		},
	})
	if err != nil {
		return fmt.Errorf("notes roomId+category+createdAt index: %w", err)
	}

	return nil
}

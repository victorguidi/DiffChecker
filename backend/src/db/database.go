package db

import (
	"log"
	"time"

	"context"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	Database   *mongo.Database
	Collection *mongo.Collection
}

func NewDatabase() (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set client options
	opt := options.Client().SetMaxPoolSize(10).ApplyURI(os.Getenv("MONGO_URI"))

	client, err := mongo.Connect(ctx, opt)
	if err != nil {
		return nil, err
	}

	return &DB{
		Database:   client.Database("DiffChecker"),
		Collection: client.Database("DiffChecker").Collection("diffs"),
	}, nil
}

func (d *DB) FindDiffBy(query bson.D, obj interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := d.Collection.FindOne(ctx, query).Decode(obj)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) FindAll(obj interface{}, query bson.D) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := d.Collection.Find(ctx, query)
	if err != nil {
		return err
	}
	if err = cursor.All(ctx, obj); err != nil {
		return err
	}
	return nil
}

func (d *DB) InsertDiff(diff interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := d.Collection.InsertOne(ctx, diff)
	if err != nil {
		log.Default().Printf("Error inserting diff: %v", err)
	}
	return nil
}

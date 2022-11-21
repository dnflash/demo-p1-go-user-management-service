package database

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	UserDB          = "userDB"
	CollectionUsers = "users"
)

var ErrNoDocumentsModified = errors.New("no documents modified")

type UserDatabase struct {
	*mongo.Database
}

func ConnectUserDB(ctx context.Context, dbURI string) (*mongo.Client, error) {
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURI))
	if err != nil {
		return nil, err
	}

	_, err = c.Database(UserDB).Collection(CollectionUsers).Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}

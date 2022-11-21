package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	Username string             `bson:"username" json:"username"`
	Password []byte             `bson:"password" json:"-"`
	Role     string             `bson:"role" json:"role"`
	Info     string             `bson:"info" json:"info"`
}

func (db UserDatabase) InsertUser(ctx context.Context, u User) (string, error) {
	r, err := db.Collection(CollectionUsers).InsertOne(ctx, u)
	if err != nil {
		return "", fmt.Errorf("error inserting User with username: %v: %w", u.Username, err)
	}
	return r.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (db UserDatabase) FindUserByID(ctx context.Context, id string) (User, error) {
	var u User
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return u, fmt.Errorf("error creating ObjectID from hex: %s: %w", id, err)
	}
	err = db.Collection(CollectionUsers).FindOne(ctx, bson.M{"_id": objID}).Decode(&u)
	if err != nil {
		return u, fmt.Errorf("error finding User with ID: %s: %w", id, err)
	}
	return u, nil
}

func (db UserDatabase) FindUserByUsername(ctx context.Context, username string) (User, error) {
	var u User
	err := db.Collection(CollectionUsers).FindOne(ctx, bson.M{"username": username}).Decode(&u)
	if err != nil {
		return u, fmt.Errorf("error finding User with username: %s: %w", username, err)
	}
	return u, nil
}

func (db UserDatabase) FindAllUsers(ctx context.Context) ([]User, error) {
	var us []User
	cur, err := db.Collection(CollectionUsers).Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("error getting cursor to find all Users: %w", err)
	}
	if err = cur.All(ctx, &us); err != nil {
		return nil, fmt.Errorf("error getting all Users from cursor: %w", err)
	}
	return us, nil
}

func (db UserDatabase) UpdateUserPassword(ctx context.Context, username string, password []byte) error {
	r, err := db.Collection(CollectionUsers).UpdateOne(ctx,
		bson.M{"username": username},
		bson.M{"$set": bson.M{"password": password}},
	)
	if err != nil {
		return fmt.Errorf("error updating User password, username: %v, err: %w", username, err)
	}
	if r.ModifiedCount == 0 {
		return fmt.Errorf("no documents modified when updating user password, username: %v, err: %w", username, ErrNoDocumentsModified)
	}
	return nil
}

func (db UserDatabase) UpdateUserInfo(ctx context.Context, username string, info string) error {
	r, err := db.Collection(CollectionUsers).UpdateOne(ctx,
		bson.M{"username": username},
		bson.M{"$set": bson.M{"info": info}},
	)
	if err != nil {
		return fmt.Errorf("error updating User info, username: %v, info: %v, err: %w", username, info, err)
	}
	if r.ModifiedCount == 0 {
		return fmt.Errorf("no documents modified when updating user info, username: %v, info: %v, err: %w", username, info, ErrNoDocumentsModified)
	}
	return nil
}

func (db UserDatabase) UpdateUserRole(ctx context.Context, username string, role string) error {
	r, err := db.Collection(CollectionUsers).UpdateOne(ctx,
		bson.M{"username": username},
		bson.M{"$set": bson.M{"role": role}},
	)
	if err != nil {
		return fmt.Errorf("error updating User role, username: %v, role: %v, err: %w", username, role, err)
	}
	if r.ModifiedCount == 0 {
		return fmt.Errorf("no documents modified when updating user role, username: %v, role: %v, err: %w", username, role, ErrNoDocumentsModified)
	}
	return nil
}

func (db UserDatabase) DeleteUserByUsername(ctx context.Context, username string) error {
	_, err := db.Collection(CollectionUsers).DeleteOne(ctx, bson.M{"username": username})
	if err != nil {
		return fmt.Errorf("error deleting User with username: %s: %w", username, err)
	}
	return nil
}

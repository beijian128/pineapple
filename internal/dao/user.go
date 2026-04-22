
package dao

import (
	"context"
	"time"

	"github.com/beijian128/pineapple/internal/model"
	"github.com/beijian128/pineapple/internal/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	UserCollection = "users"
)

type UserDAO struct {
	collection *mongo.Collection
}

func NewUserDAO() *UserDAO {
	return &UserDAO{
		collection: storage.GetCollection(UserCollection),
	}
}

func (dao *UserDAO) Create(user *model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := dao.collection.InsertOne(ctx, user)
	return err
}

func (dao *UserDAO) FindByUsername(username string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user model.User
	err := dao.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (dao *UserDAO) FindByID(id primitive.ObjectID) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user model.User
	err := dao.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (dao *UserDAO) UpdateLastLogin(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"last_login_at": now,
			"updated_at":    now,
		},
	}

	_, err := dao.collection.UpdateByID(ctx, id, update)
	return err
}

func (dao *UserDAO) ExistsUsername(username string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := dao.collection.CountDocuments(ctx, bson.M{"username": username})
	return count > 0, err
}

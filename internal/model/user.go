
package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username     string             `bson:"username" json:"username"`
	Password     string             `bson:"password" json:"-"`
	Nickname     string             `bson:"nickname" json:"nickname"`
	Email        string             `bson:"email,omitempty" json:"email"`
	Phone        string             `bson:"phone,omitempty" json:"phone"`
	Avatar       string             `bson:"avatar,omitempty" json:"avatar"`
	Status       int                `bson:"status" json:"status"`
	LastLoginAt  time.Time          `bson:"last_login_at,omitempty" json:"last_login_at"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

const (
	UserStatusNormal  = 1
	UserStatusBanned  = 2
	UserStatusDeleted = 3
)

func NewUser(username, password string) *User {
	now := time.Now()
	return &User{
		ID:         primitive.NewObjectID(),
		Username:   username,
		Password:   password,
		Nickname:   username,
		Status:     UserStatusNormal,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}


package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beijian128/pineapple/internal/model"
	"github.com/beijian128/pineapple/internal/storage"
	"github.com/beijian128/pineapple/internal/utils"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	SessionPrefix = "session:"
	SessionTTL    = 7 * 24 * time.Hour
)

type Session struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Nickname string    `json:"nickname"`
	ExpireAt time.Time `json:"expire_at"`
}

func CreateSession(user *model.User) (string, error) {
	token, err := utils.GenerateToken()
	if err != nil {
		return "", err
	}

	session := &Session{
		UserID:   user.ID.Hex(),
		Username: user.Username,
		Nickname: user.Nickname,
		ExpireAt: time.Now().Add(SessionTTL),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return "", err
	}

	key := SessionPrefix + token
	if storage.RedisClient != nil {
		ctx := context.Background()
		err = storage.RedisClient.Set(ctx, key, data, SessionTTL).Err()
		if err != nil {
			utils.Logger.Warn("redis session set failed", zap.Error(err))
		}
	}

	return token, nil
}

func GetSession(token string) (*Session, error) {
	if storage.RedisClient == nil {
		return nil, fmt.Errorf("redis not initialized")
	}

	key := SessionPrefix + token
	ctx := context.Background()
	data, err := storage.RedisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var session Session
	err = json.Unmarshal([]byte(data), &session)
	if err != nil {
		return nil, err
	}

	if time.Now().After(session.ExpireAt) {
		return nil, nil
	}

	return &session, nil
}

func DeleteSession(token string) error {
	if storage.RedisClient == nil {
		return fmt.Errorf("redis not initialized")
	}

	key := SessionPrefix + token
	ctx := context.Background()
	return storage.RedisClient.Del(ctx, key).Err()
}

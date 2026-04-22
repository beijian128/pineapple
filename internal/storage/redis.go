
package storage

import (
	"context"
	"fmt"

	"github.com/beijian128/pineapple/internal/utils"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var RedisClient *redis.Client

func InitRedis(cfg *utils.RedisConfig) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx := context.Background()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	utils.Logger.Info("redis connected successfully",
		zap.String("addr", cfg.Addr))

	return nil
}

func CloseRedis() {
	if RedisClient != nil {
		_ = RedisClient.Close()
		utils.Logger.Info("redis disconnected")
	}
}

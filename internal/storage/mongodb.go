
package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/beijian128/pineapple/internal/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var MongoClient *mongo.Client
var MongoDatabase *mongo.Database

func InitMongoDB(cfg *utils.MongoDBConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(cfg.URI)
	if cfg.MaxPoolSize > 0 {
		clientOpts.SetMaxPoolSize(cfg.MaxPoolSize)
	}
	if cfg.MinPoolSize > 0 {
		clientOpts.SetMinPoolSize(cfg.MinPoolSize)
	}

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping mongodb: %w", err)
	}

	MongoClient = client
	MongoDatabase = client.Database(cfg.Database)

	utils.Logger.Info("mongodb connected successfully",
		zap.String("database", cfg.Database))

	return nil
}

func CloseMongoDB() {
	if MongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = MongoClient.Disconnect(ctx)
		utils.Logger.Info("mongodb disconnected")
	}
}

func GetCollection(name string) *mongo.Collection {
	if MongoDatabase == nil {
		return nil
	}
	return MongoDatabase.Collection(name)
}

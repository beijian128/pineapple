
package utils

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Network NetworkConfig `mapstructure:"network"`
	Etcd    EtcdConfig    `mapstructure:"etcd"`
	MongoDB MongoDBConfig `mapstructure:"mongodb"`
	Redis   RedisConfig   `mapstructure:"redis"`
	Log     LogConfig     `mapstructure:"log"`
}

type ServerConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Env     string `mapstructure:"env"`
}

type NetworkConfig struct {
	TCP  TCPConfig  `mapstructure:"tcp"`
	WS   WSConfig   `mapstructure:"websocket"`
	KCP  KCPConfig  `mapstructure:"kcp"`
	GRPC GRPCConfig `mapstructure:"grpc"`
}

type TCPConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
}

type WSConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
}

type KCPConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
}

type GRPCConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
}

type EtcdConfig struct {
	Endpoints   []string `mapstructure:"endpoints"`
	DialTimeout string   `mapstructure:"dial_timeout"`
	LeaseTTL    int64    `mapstructure:"lease_ttl"`
}

type MongoDBConfig struct {
	URI         string `mapstructure:"uri"`
	Database    string `mapstructure:"database"`
	MaxPoolSize uint64 `mapstructure:"max_pool_size"`
	MinPoolSize uint64 `mapstructure:"min_pool_size"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

var AppConfig *Config

func LoadConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

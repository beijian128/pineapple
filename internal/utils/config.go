package utils

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Etcd EtcdConfig `mapstructure:"etcd"`
	Log  LogConfig  `mapstructure:"log"`
}

type EtcdConfig struct {
	Endpoints   []string `mapstructure:"endpoints"`
	DialTimeout string   `mapstructure:"dial_timeout"`
	LeaseTTL    int64    `mapstructure:"lease_ttl"`
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

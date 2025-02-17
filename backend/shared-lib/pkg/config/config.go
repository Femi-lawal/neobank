package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

func LoadConfig(path string, config interface{}) error {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: config file not found, using defaults/env vars: %v", err)
	}

	if err := viper.Unmarshal(config); err != nil {
		return err
	}

	return nil
}

package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// LoadBasicConfig loads configuration from a YAML file into the provided config struct.
// This is a simplified config loader for basic use cases.
// For AWS-integrated configuration with secrets, use LoadServiceConfig instead.
func LoadBasicConfig(path string, config interface{}) error {
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

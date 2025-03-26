package config

import (
	"github.com/spf13/viper"
)

// Config holds all application settings
type Config struct {
	ServerPort string `mapstructure:"SERVER_PORT"`
	MongoURI   string `mapstructure:"MONGO_URI"`
	MongoDB    string `mapstructure:"MONGO_DB"`
	LogLevel   string `mapstructure:"LOG_LEVEL"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	viper.SetDefault("SERVER_PORT", "50051")
	viper.SetDefault("MONGO_URI", "mongodb://localhost:27017")
	viper.SetDefault("MONGO_DB", "barbershop_bookings")
	viper.SetDefault("LOG_LEVEL", "info")

	viper.AutomaticEnv()

	config := &Config{
		ServerPort: viper.GetString("SERVER_PORT"),
		MongoURI:   viper.GetString("MONGO_URI"),
		MongoDB:    viper.GetString("MONGO_DB"),
		LogLevel:   viper.GetString("LOG_LEVEL"),
	}

	return config, nil
}

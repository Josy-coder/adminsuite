package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"

)

type Config struct {
    DBHost           string `mapstructure:"DB_HOST"`
    DBPort           string `mapstructure:"DB_PORT"`
    DBUser           string `mapstructure:"DB_USER"`
    DBPassword       string `mapstructure:"DB_PASSWORD"`
    DBName           string `mapstructure:"DB_NAME"`
    ServerPort       string `mapstructure:"SERVER_PORT"`
    PasetoPublicKey  string `mapstructure:"PASETO_PUBLIC_KEY"`
    PasetoPrivateKey string `mapstructure:"PASETO_PRIVATE_KEY"`
	PasetoKey string

    SMTPHost     string
    SMTPPort     int
    SMTPUsername string
    SMTPPassword string
    SMTPFrom     string

    TwilioAccountSID  string
    TwilioAuthToken   string
    TwilioPhoneNumber string
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No config file found, using environment variables")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return &config, nil
}
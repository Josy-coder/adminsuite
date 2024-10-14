package config

import (
	"github.com/spf13/viper"

)

type Config struct {
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	ServerPort string `mapstructure:"SERVER_PORT"`

	PasetoPublicKey  string `mapstructure:"PASETO_PUBLIC_KEY"`
	PasetoPrivateKey string `mapstructure:"PASETO_PRIVATE_KEY"`
	PasetoKey        []byte

	SMTPHost     string `mapstructure:"SMTP_HOST"`
	SMTPPort     int    `mapstructure:"SMTP_PORT"`
	SMTPUsername string `mapstructure:"SMTP_USERNAME"`
	SMTPPassword string `mapstructure:"SMTP_PASSWORD"`
	SMTPFrom     string `mapstructure:"SMTP_FROM"`

	TwilioAccountSID  string `mapstructure:"TWILIO_ACCOUNT_SID"`
	TwilioAuthToken   string `mapstructure:"TWILIO_AUTH_TOKEN"`
	TwilioPhoneNumber string `mapstructure:"TWILIO_PHONE_NUMBER"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "")
	viper.SetDefault("DB_NAME", "adminsuitedb")
	viper.SetDefault("SERVER_PORT", "8080")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}


	config.PasetoKey = []byte(config.PasetoPrivateKey)

	return &config, nil
}

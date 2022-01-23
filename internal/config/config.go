package config

import (
	"os"
)

type Config struct {
	ApplicationPort string
	DBAuthenticationData
	APIData
}

type APIData struct {
	Key  string
	URL  string
	Path string
}

type DBAuthenticationData struct {
	DBAdminUsername string
	DBAdminPassword string
	DBName          string
	DBHost          string
	DBPort          string
	URI             string
}

func New() (*Config, error) {
	return &Config{
		ApplicationPort: os.Getenv("PORT"),
		DBAuthenticationData: DBAuthenticationData{
			DBAdminUsername: os.Getenv("DB_ADMIN_USERNAME"),
			DBAdminPassword: os.Getenv("DB_ADMIN_PASSWORD"),
			DBHost:          os.Getenv("DB_HOST"),
			DBPort:          os.Getenv("DB_PORT"),
			DBName:          os.Getenv("DB_NAME"),
			URI:             os.Getenv("POSTGRES_URI"),
		},
		APIData: APIData{
			Key:  os.Getenv("EXCHANGE_API_KEY"),
			URL:  os.Getenv("EXCHANGE_API_URL"),
			Path: os.Getenv("EXCHANGE_API_PATH"),
		},
	}, nil
}

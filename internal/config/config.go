package config

import "os"

type Config struct {
	MongoURI   string
	JWTSecret  string
	Port       string
}

func Load() *Config {
	return &Config{
		MongoURI:  getEnv("MONGO_URI", "mongodb://localhost:27017/healthy_pay"),
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key"),
		Port:      getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Copyright 2026 go-ordersvc Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	Cache    CacheConfig
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name        string
	Version     string
	Environment string
	LogLevel    string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	HTTPPort        int
	GRPCPort        int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	EnablePprof     bool
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	MigrationsPath  string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host        string
	Port        int
	Password    string
	DB          int
	MaxRetries  int
	PoolSize    int
	PoolTimeout time.Duration
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	DefaultTTL time.Duration
	HotTTL     time.Duration
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	return &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "ordersvc"),
			Version:     getEnv("APP_VERSION", "dev"),
			Environment: getEnv("APP_ENVIRONMENT", "development"),
			LogLevel:    getEnv("APP_LOG_LEVEL", "info"),
		},
		Server: ServerConfig{
			HTTPPort:        getEnvAsInt("HTTP_PORT", 8080),
			GRPCPort:        getEnvAsInt("GRPC_PORT", 9090),
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 30 * time.Second,
			EnablePprof:     false,
		},
		Database: DatabaseConfig{
			Host:            getEnv("DATABASE_HOST", "localhost"),
			Port:            getEnvAsInt("DATABASE_PORT", 5432),
			User:            getEnv("DATABASE_USER", "postgres"),
			Password:        getEnv("DATABASE_PASSWORD", "postgres"),
			Database:        getEnv("DATABASE_NAME", "ordersvc"),
			SSLMode:         getEnv("DATABASE_SSL_MODE", "disable"),
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 10 * time.Minute,
			MigrationsPath:  "file://db/migrations",
		},
		Redis: RedisConfig{
			Host:        getEnv("REDIS_HOST", "localhost"),
			Port:        getEnvAsInt("REDIS_PORT", 6379),
			Password:    getEnv("REDIS_PASSWORD", ""),
			DB:          getEnvAsInt("REDIS_DB", 0),
			MaxRetries:  3,
			PoolSize:    10,
			PoolTimeout: 4 * time.Second,
		},
		Kafka: KafkaConfig{
			Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			Topic:   getEnv("KAFKA_TOPIC", "order-events"),
			GroupID: getEnv("KAFKA_GROUP_ID", "ordersvc"),
		},
		Cache: CacheConfig{
			DefaultTTL: 5 * time.Minute,
			HotTTL:     1 * time.Hour,
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

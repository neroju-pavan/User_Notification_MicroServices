package config

import "fmt"

type Config struct {
	Listen   string   `koanf:"listen"`
	Postgres Postgres `koanf:"postgres"`
	Redis    Redis    `koanf:"redis"`

	Kafka Kafka `koanf:"kafka"`
}

type Postgres struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	User     string `koanf:"user"`
	Password string `koanf:"password"`
	Dbname   string `koanf:"db"`
}

type Redis struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Password string `koanf:"password"`
	DB       int    `koanf:"db"`
}

type Kafka struct {
	Brokers []string `koanf:"brokers"`
	Topic   string   `koanf:"topic"`

	ProducerGroupID string `koanf:"producer_group"`
}

func (c *Config) Validate() error {
	// server
	if c.Listen == "" {
		return fmt.Errorf("listen address is required")
	}

	// postgres
	if c.Postgres.Host == "" {
		return fmt.Errorf("postgres host is required")
	}
	if c.Postgres.Port == 0 {
		return fmt.Errorf("postgres port is required")
	}
	if c.Postgres.User == "" {
		return fmt.Errorf("postgres user is required")
	}
	if c.Postgres.Password == "" {
		return fmt.Errorf("postgres password is required")
	}
	if c.Postgres.Dbname == "" {
		return fmt.Errorf("postgres dbname is required")
	}

	// redis
	if c.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	if c.Redis.Port == 0 {
		return fmt.Errorf("redis port is required")
	}
	// password can be empty (public redis)
	// db can be 0 → valid

	return nil
}

var DefaultConfig = Config{
	Listen: "localhost:8083",

	Postgres: Postgres{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "pavan",
		Dbname:   "gotest",
	},

	Redis: Redis{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
	},
	Kafka: Kafka{
		Brokers:         []string{"localhost:9092"},
		Topic:           "email-service",
		ProducerGroupID: "notify-producer",
	},
}

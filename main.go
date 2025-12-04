package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"test123/config"
	"test123/http"
	kafka "test123/kafka/producers"

	"test123/repositories"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/willf/bloom"
)

func warmBloomFilter(bf *bloom.BloomFilter, db *pgxpool.Pool) {
	rows, err := db.Query(context.Background(), "SELECT username FROM users")
	if err != nil {
		fmt.Println("Bloom warm failed:", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err == nil {
			bf.AddString(username)
			count++
		}
	}

	//fmt.Println("🔥 Bloom Filter warmed with", count, "usernames")
}
func InitializeServer(cfg config.Config, ctx context.Context) (
	*http.Server,
	*pgxpool.Pool,
	*redis.Client,
	*kafka.KafkaNotificationProducer,

	error,
) {

	pool, err := repositories.Connect(ctx, cfg)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: 10,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, nil, nil, nil, err
	}
	bloomFilter := bloom.NewWithEstimates(1_000_000, 0.01)
	go warmBloomFilter(bloomFilter, pool)
	producer := kafka.NewKafkaNotificationProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	LoadEnv()
	appServer := http.NewServer("Connected", pool, rdb, producer, bloomFilter)
	return appServer, pool, rdb, producer, nil
}

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.DefaultConfig

	server, db, rdb, prod, err := InitializeServer(cfg, ctx)
	if err != nil {
		panic(" Failed to initialize server: " + err.Error())
	}

	defer db.Close()
	defer rdb.Close()
	defer prod.Writer.Close()

	fmt.Println(" Starting HTTP Server on", cfg.Listen)
	if err := server.Listen(ctx, cfg.Listen); err != nil {
		panic(" Server error: " + err.Error())
	}
}

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}
}

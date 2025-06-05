package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"main/internal/model"
	dbP "main/pkg/db/postgres"
	oredis "main/pkg/redis"
	"main/router"
	"time"
)

var ctx = context.Background()

func main() {
	d := ConnectPostgres()
	ConnectRedis()
	dbP.SetCluster(d)
	engine := gin.Default()
	//gin.SetMode(gin.ReleaseMode)
	router.Internal(ctx, engine)

	engine.Run(":8080")
}

func ConnectPostgres() *gorm.DB {
	dsn := "host=localhost user=admin password=admin dbname=crud port=5433 sslmode=disable TimeZone=Asia/Kolkata"

	gormConfig := gorm.Config{}
	gormConfig.Logger = logger.Default.LogMode(logger.Info)

	db, err := gorm.Open(postgres.Open(dsn), &gormConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate your models
	err = db.AutoMigrate(&model.Campaign{})
	if err != nil {
		log.Fatalf("Auto migration failed: %v", err)
	}

	fmt.Println("Connected to PostgreSQL & migrated successfully")
	return db
}

func ConnectRedis() {
	rdb := redis.NewClient(&redis.Options{
		Addr:         "localhost:7005",
		Password:     "",
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	fmt.Println("Redis PING response:", pong) // Should print: PONG

	oredis.SetClient(rdb)
}

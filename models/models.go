package models

import (
	"content-clock/helpers"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var Redis *redis.Client
var ctx = context.Background()

func ConnectDatabase(dsn_url, env string) {

	// Configure logger
	var logLevel logger.LogLevel
	if env == "prod" {
		logLevel = logger.Error
	} else {
		logLevel = logger.Info
	}

	// Set the logger configuration
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logLevel,    // Log level
			Colorful:      true,        // Enable color
		},
	)

	database, db_err := gorm.Open(postgres.Open(dsn_url), &gorm.Config{
		Logger: newLogger,
	})
	// database, db_err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if db_err != nil {
		helpers.Logging("error", db_err.Error())
		return
	}
	db_migrate := os.Getenv("DB_MIGRATE")
	if db_migrate == "true" {
		err := database.AutoMigrate(&Connections{}, &SocialPosts{})
		if err != nil {
			helpers.Logging("error", err.Error())
			return
		}
	}

	getDb, err := database.DB()
	if err != nil {
		helpers.Logging("error", err.Error())
		return
	}
	getDb.SetMaxIdleConns(10)
	getDb.SetMaxOpenConns(100)
	getDb.SetConnMaxLifetime(time.Hour)
	DB = database
}

func ConnectRedis(host, port, user, password, db, env string) {
	// db str to int
	dbInt, _ := strconv.Atoi(db)
	options := &redis.Options{
		Addr:        fmt.Sprintf("%s:%s", host, port),
		Password:    password, // no password set
		DB:          dbInt,    // use default DB
		Username:    user,
		ReadTimeout: -1,
	}

	// Apply TLS configuration if environment is "prod"
	if env == "prod" {
		options.TLSConfig = &tls.Config{
			ServerName: host,
		}
	}

	// Initialize Redis client
	rdb := redis.NewClient(options)

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Error connecting to Redis:", err)
	}

	Redis = rdb
}

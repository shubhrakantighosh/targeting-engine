package init

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	config "github.com/spf13/viper"
	opostgres "main/pkg/db/postgres"
	oredis "main/pkg/redis"
	"strings"
	"time"
)

func Initialize(ctx context.Context) {
	initializeRedis(ctx)
	initializeDB()
}

func initializeRedis(ctx context.Context) {
	r := redis.NewClient(&redis.Options{
		Addr: config.GetString("redis.host"),
		DB:   config.GetInt("redis.db"),
	})
	fmt.Println("Initialized Redis Client")
	oredis.SetClient(r)
}

func initializeDB() {
	maxOpenConnections := config.GetInt("postgresql.maxOpenConns")
	maxIdleConnections := config.GetInt("postgresql.maxIdleConns")

	database := config.GetString("postgresql.database")
	connIdleTimeout := 10 * time.Minute

	// Read Write endpoint config
	mysqlWriteServer := config.GetString("postgresql.master.host")
	mysqlWritePort := config.GetString("postgresql.master.port")
	mysqlWritePassword := config.GetString("postgresql.master.password")
	mysqlWriterUsername := config.GetString("postgresql.master.username")

	// Fetch Read endpoint config
	mysqlReadServers := config.GetString("postgresql.slaves.hosts")
	mysqlReadPort := config.GetString("postgresql.slaves.port")
	mysqlReadPassword := config.GetString("postgresql.slaves.password")
	mysqlReadUsername := config.GetString("postgresql.slaves.username")

	debugMode := config.GetBool("postgresql.debugMode")

	// Master config
	masterConfig := opostgres.DBConfig{
		Host:               mysqlWriteServer,
		Port:               mysqlWritePort,
		Username:           mysqlWriterUsername,
		Password:           mysqlWritePassword,
		Dbname:             database,
		MaxOpenConnections: maxOpenConnections,
		MaxIdleConnections: maxIdleConnections,
		ConnMaxLifetime:    connIdleTimeout,
		DebugMode:          debugMode,
	}

	slavesConfig := make([]opostgres.DBConfig, 0)
	for _, host := range strings.Split(mysqlReadServers, ",") {
		slaveConfig := opostgres.DBConfig{
			Host:               host,
			Port:               mysqlReadPort,
			Username:           mysqlReadUsername,
			Password:           mysqlReadPassword,
			Dbname:             database,
			MaxOpenConnections: maxOpenConnections,
			MaxIdleConnections: maxIdleConnections,
			ConnMaxLifetime:    connIdleTimeout,
			DebugMode:          debugMode,
		}
		slavesConfig = append(slavesConfig, slaveConfig)
	}

	db := opostgres.InitializeDBInstance(masterConfig, &slavesConfig)
	fmt.Println("Initialized Postgres DB client")
	opostgres.SetCluster(db)
}

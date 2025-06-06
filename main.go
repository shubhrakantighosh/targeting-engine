package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"main/config"
	initilizer "main/init"
	"main/router"
)

func main() {
	ctx := context.TODO()

	config.InitConfig()
	initilizer.Initialize(ctx)

	// TODO seperated http func init
	app := gin.New()
	router.Internal(ctx, app)

	port := viper.GetString("server.port")
	if err := app.Run(port); err != nil {
		panic("failed to start server: " + err.Error())
	}
}

package router

import (
	"context"
	"github.com/gin-gonic/gin"
	campaignController "main/internal/controller/campaign"
	"main/middlewares"
	"main/pkg/db/postgres"
	"main/pkg/redis"
)

func Internal(ctx context.Context, s *gin.Engine) {

	g := s.Group("internal/api/v1")

	delivery := g.Group("/delivery")
	{

		ser := campaignController.Wire(ctx, postgres.GetCluster().DbCluster, redis.GetClient().Client)

		delivery.GET("", middlewares.SanitizeQueryParams(), ser.GetAll)
	}

}

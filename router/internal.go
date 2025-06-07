package router

import (
	"context"
	"github.com/gin-gonic/gin"
	deliveryController "main/internal/controller/delivery"
	deliveryService "main/internal/delivery/service"
	"main/middlewares"
	"main/pkg/db/postgres"
	"main/pkg/redis"
)

func Internal(ctx context.Context, s *gin.Engine) {
	targetingEngineV1 := s.Group("internal/api/v1")

	deliveryRoute := targetingEngineV1.Group("/delivery")
	{
		deliverySvc := deliveryService.Wire(ctx, postgres.GetCluster().DbCluster, redis.GetClient())
		deliveryValidator := middlewares.NewDeliveryValidator(deliverySvc)

		deliveryCtrl := deliveryController.Wire(ctx, postgres.GetCluster().DbCluster, redis.GetClient())
		deliveryRoute.GET("",
			middlewares.SanitizeQueryParams(), deliveryValidator.ValidateApp(),
			deliveryCtrl.GetMatchingCampaigns,
		)
	}
}

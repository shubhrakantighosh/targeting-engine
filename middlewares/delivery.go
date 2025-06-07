package middlewares

import (
	"github.com/gin-gonic/gin"
	"main/constants"
	deliveryService "main/internal/delivery/service"
	"net/http"
	"sync"
)

type DeliveryValidator struct {
	deliveryService.Interface
}

var (
	syncOnce  sync.Once
	validator *DeliveryValidator
)

func NewDeliveryValidator(svc deliveryService.Interface) *DeliveryValidator {
	syncOnce.Do(func() {
		validator = &DeliveryValidator{svc}
	})

	return validator
}

// change to app to dynamic so anyone can use this func instaed of hradcore app
func (v *DeliveryValidator) ValidateApp() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		app := ctx.Query(constants.App)
		if len(app) == 0 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "app required"})
			return
		}

		found, cusErr := v.AppExists(ctx, app)
		if cusErr.Exists() {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Please try again"})
			return
		}

		if !found {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Please provide a valid app"})
			return
		}

		ctx.Next()
	}
}

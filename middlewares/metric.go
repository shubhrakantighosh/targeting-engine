package middlewares

import (
	"github.com/gin-gonic/gin"
	"main/constants"
	opromauto "main/pkg/metric/prometheus"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		app := ctx.Query(constants.App)
		ctx.Next() // Process request

		if len(app) == 0 {
			app = "unknown"
		}
		// After request
		statusCode := ctx.Writer.Status()
		method := ctx.Request.Method
		opromauto.GetMetrics().RequestCounter.WithLabelValues(app, method, toStatusClass(statusCode)).Inc()
	}
}

func toStatusClass(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}

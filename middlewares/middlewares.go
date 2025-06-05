package middlewares

import (
	"github.com/gin-gonic/gin"
	"main/utils"
)

func SanitizeQueryParams() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		queryParams := ctx.Request.URL.Query()
		for k, v := range queryParams {
			queryParams[utils.TrimSpace(k)] = utils.TrimStrings(v)
		}

		ctx.Next()
	}
}

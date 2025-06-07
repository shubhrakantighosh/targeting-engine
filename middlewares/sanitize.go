package middlewares

import (
	"github.com/gin-gonic/gin"
	"main/util"
	"net/url"
)

func SanitizeQueryParams() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		queryParams := make(url.Values)
		for k, v := range ctx.Request.URL.Query() {
			queryParams[util.TrimSpace(k)] = util.TrimStrings(v)
		}

		ctx.Request.URL.RawQuery = queryParams.Encode()
		ctx.Next()
	}
}

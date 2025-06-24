package middlewares

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"main/internal/controller/delivery/request"
	"main/pkg/apperror"
	ocache "main/pkg/cache"
	"net/http"
)

func TargetingCacheMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var params request.DeliveryRequestParams
		if err := ctx.ShouldBindQuery(&params); err != nil {
			apperror.New(err, http.StatusBadRequest).AbortWithError(ctx)
			return
		}

		key := cacheKey(params)

		if data, found := ocache.GetClient().Get(key); found {
			ctx.JSON(http.StatusOK, data)
			ctx.Abort()
			return
		}

		writer := &responseCaptureWriter{ResponseWriter: ctx.Writer, body: make([]byte, 0)}
		ctx.Writer = writer

		ctx.Next()

		if ctx.Writer.Status() == http.StatusOK {
			var responseData interface{}
			if err := json.Unmarshal(writer.body, &responseData); err == nil {
				ocache.GetClient().Set(key, responseData, cache.DefaultExpiration)
			}
		}
	}
}

type responseCaptureWriter struct {
	gin.ResponseWriter
	body []byte
}

func (w *responseCaptureWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.ResponseWriter.Write(data)
}

func cacheKey(params request.DeliveryRequestParams) string {
	return params.App + "|" + params.Country + "|" + params.OS + "|" + params.State + "|" + params.City
}

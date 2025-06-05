package campaign

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (ctrl *Controller) GetAll(ctx *gin.Context) {
	d, _, err := ctrl.service.GetAll(ctx, map[string]any{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": d,
	})
}

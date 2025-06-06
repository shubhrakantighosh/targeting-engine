package delivery

import (
	"github.com/gin-gonic/gin"
	"main/internal/controller/delivery/adapter"
	"main/internal/controller/delivery/request"
	"main/pkg/apperror"
	"net/http"
)

func (ctrl *Controller) GetMatchingCampaigns(ctx *gin.Context) {
	var deliveryRequestParams request.DeliveryRequestParams
	if err := ctx.ShouldBindQuery(&deliveryRequestParams); err != nil {
		apperror.New(err, http.StatusBadRequest).AbortWithError(ctx)
		return
	}

	if cusErr := deliveryRequestParams.Validate(); cusErr.Exists() {
		cusErr.AbortWithError(ctx)
		return
	}

	campaigns, cusErr := ctrl.service.GetMatchingCampaigns(ctx, deliveryRequestParams)
	if cusErr.Exists() {
		cusErr.AbortWithError(ctx)
		return
	}

	if len(campaigns) == 0 {
		ctx.AbortWithStatusJSON(http.StatusNoContent, nil)
		return
	}

	ctx.AbortWithStatusJSON(http.StatusOK, adapter.TransformCampaignsForController(campaigns))
	return
}

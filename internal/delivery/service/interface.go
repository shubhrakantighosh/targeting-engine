package service

import (
	"context"
	"main/internal/controller/delivery/request"
	"main/internal/model"
	"main/pkg/apperror"
)

type Interface interface {
	GetMatchingCampaigns(
		ctx context.Context,
		deliveryRequestParams request.DeliveryRequestParams,
	) (model.Campaigns, apperror.Error)

	AppExists(
		ctx context.Context,
		app string,
	) (isExists bool, cusErr apperror.Error)
}

package service

import (
	"context"
	"main/internal/controller/delivery/request"
	"main/internal/model"
	"main/pkg/errors"
)

type Interface interface {
	GetMatchingCampaigns(
		ctx context.Context,
		deliveryRequestParams request.DeliveryRequestParams,
	) (model.Campaigns, errors.Error)
}

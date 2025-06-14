package request

import (
	"main/internal/model"
	"main/internal/targeting_rule/service"
	"main/pkg/apperror"
	"net/http"
)

type DeliveryRequestParams struct {
	App     string `form:"app" binding:"required"`
	Country string `form:"country" binding:"required"`
	OS      string `form:"os" binding:"required"`
	State   string `form:"state"`
	City    string `form:"city"`
}

func (q DeliveryRequestParams) Validate() (err apperror.Error) {
	if len(q.App) == 0 {
		err = apperror.NewWithMessage("app parameter is required", http.StatusBadRequest)
		return
	}

	if len(q.Country) == 0 {
		err = apperror.NewWithMessage("country parameter is required", http.StatusBadRequest)
		return
	}

	if len(q.OS) == 0 {
		err = apperror.NewWithMessage("os parameter is required", http.StatusBadRequest)
		return
	}

	return
}

func (q DeliveryRequestParams) DimensionTypeMapValue() (dimensionTypeMapValue service.ClientRequest) {
	dimensionTypeMapValue = make(service.ClientRequest)
	dimensionTypeMapValue[model.App] = q.App
	dimensionTypeMapValue[model.Country] = q.Country
	dimensionTypeMapValue[model.OS] = q.OS

	if len(q.State) > 0 {
		dimensionTypeMapValue[model.State] = q.State
	}

	if len(q.City) > 0 {
		dimensionTypeMapValue[model.City] = q.City
	}

	return
}

func (q DeliveryRequestParams) ToBuildHierarchy() (dimensionType []model.DimensionType) {
	dimensionType = make([]model.DimensionType, 0)
	if len(q.Country) == 0 || len(q.State) == 0 {
		return
	}

	if len(q.Country) > 0 && len(q.State) > 0 {
		dimensionType = append(dimensionType, model.Country, model.State)
	}

	if len(q.State) > 0 && len(q.City) > 0 {
		dimensionType = append(dimensionType, model.City)
	}

	return
}

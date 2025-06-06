package request

import (
	"main/constants"
	"main/pkg/errors"
	"net/http"
	"net/url"
)

type DeliveryRequestParams struct {
	App     string `form:"app" binding:"required"`
	Country string `form:"country" binding:"required"`
	OS      string `form:"os" binding:"required"`
}

func (q DeliveryRequestParams) Validate() (err errors.Error) {
	if len(q.App) == 0 {
		err = errors.NewErrorMsg("app parameter is required", http.StatusBadRequest)
		return
	}

	if len(q.Country) == 0 {
		err = errors.NewErrorMsg("country parameter is required", http.StatusBadRequest)
		return
	}

	if len(q.OS) == 0 {
		err = errors.NewErrorMsg("os parameter is required", http.StatusBadRequest)
		return
	}

	return
}

func (q DeliveryRequestParams) ToQueryPrams() (queryParams url.Values, err errors.Error) {
	queryParams = make(url.Values)

	if len(q.App) == 0 {
		err = errors.NewErrorMsg("app parameter is required", http.StatusBadRequest)
		return
	}
	queryParams.Add(constants.App, q.App)

	if len(q.Country) > 0 {
		queryParams.Add(constants.Country, q.Country)
	}

	if len(q.OS) > 0 {
		queryParams.Add(constants.OS, q.OS)
	}

	return
}

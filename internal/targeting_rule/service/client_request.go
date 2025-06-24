package service

import "main/internal/model"

type ClientRequest map[model.DimensionType]string

func (c ClientRequest) ToBuildHierarchy() (dimensionType []model.DimensionType) {
	dimensionType = make([]model.DimensionType, 0)
	if len(c[model.State]) == 0 {
		return
	}

	dimensionType = append(dimensionType, model.Country, model.State)

	if len(c[model.City]) > 0 {
		dimensionType = append(dimensionType, model.City)
	}

	return
}

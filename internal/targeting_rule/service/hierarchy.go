package service

import "main/internal/model"

type DimensionHierarchy struct {
	Dimensions []model.DimensionType                       // Order defines hierarchy (country -> state -> city -> district)
	ParentMap  map[model.DimensionType]model.DimensionType // child -> parent mapping
}

func BuildHierarchy(dimensions ...model.DimensionType) DimensionHierarchy {
	hierarchy := DimensionHierarchy{
		ParentMap: make(map[model.DimensionType]model.DimensionType),
	}
	if len(dimensions) > 0 {
		hierarchy.Dimensions = dimensions[:len(dimensions)-1]
	}

	for i := 0; i < len(dimensions)-1; i++ {
		hierarchy.ParentMap[dimensions[i]] = dimensions[i+1]
	}

	return hierarchy
}

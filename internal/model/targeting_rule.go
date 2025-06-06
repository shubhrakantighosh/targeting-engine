package model

import (
	"gorm.io/gorm"
	"main/util"
	"time"
)

type DimensionType string

const (
	App     DimensionType = "app"
	Country DimensionType = "country"
	OS      DimensionType = "os"
)

type TargetingRule struct {
	ID            uint64         `json:"id"`
	CampaignID    uint64         `json:"campaign_id"`
	DimensionType DimensionType  `json:"dimension_type"` // "app" "country", "os"
	Include       bool           `json:"include"`        // true = include, false = exclude
	Value         string         `json:"value"`          // actual value, e.g., "android", "us"
	CreatedBy     string         `json:"created_by"`
	UpdatedBy     string         `json:"updated_by"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at"`
}

type TargetingRules []TargetingRule

func (tr TargetingRules) IsEmpty() bool {
	return tr == nil || len(tr) == 0
}

func (tr TargetingRules) GetCampaignIDs() []uint64 {
	ids := make([]uint64, 0)
	if tr.IsEmpty() {
		return ids
	}

	for _, rule := range tr {
		ids = append(ids, rule.CampaignID)
	}

	return util.DeduplicateSlice(ids)
}

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
	State   DimensionType = "state"
	City    DimensionType = "city"
)

func (d DimensionType) String() string {
	return string(d)
}

func (d DimensionType) Is(dimensionType DimensionType) bool {
	return d == dimensionType
}

type TargetingRule struct {
	ID            uint64         `json:"id"`
	CampaignID    uint64         `json:"campaign_id"`
	ParentID      uint64         `json:"parent_id"`      // ref of TargetingRule id for state etc
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

var NonHierarchicalDimensionType = map[DimensionType]struct{}{
	App:     {},
	OS:      {},
	Country: {},
}

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

func (tr TargetingRules) GroupByCampaignID() (campaignIDMap map[uint64]TargetingRules) {
	campaignIDMap = make(map[uint64]TargetingRules)
	if tr.IsEmpty() {
		return
	}

	for _, rule := range tr {
		if _, ok := campaignIDMap[rule.CampaignID]; !ok {
			campaignIDMap[rule.CampaignID] = make(TargetingRules, 0)
		}

		campaignIDMap[rule.CampaignID] = append(campaignIDMap[rule.CampaignID], rule)
	}

	return
}

func (tr TargetingRules) HasIncludedRuleFor(dimensionType DimensionType, value string) bool {
	if tr.IsEmpty() {
		return false
	}

	for _, rule := range tr {
		if rule.Include && rule.DimensionType.Is(dimensionType) && rule.Value == value {
			return true
		}
	}

	return false
}

func (tr TargetingRules) GroupByDimensionAndValue() (grouped map[DimensionType]map[string]TargetingRules) {
	if tr.IsEmpty() {
		return
	}

	grouped = make(map[DimensionType]map[string]TargetingRules)
	for _, rule := range tr {
		if _, ok := grouped[rule.DimensionType]; !ok {
			grouped[rule.DimensionType] = make(map[string]TargetingRules)
		}
		grouped[rule.DimensionType][rule.Value] = append(grouped[rule.DimensionType][rule.Value], rule)
	}

	return
}

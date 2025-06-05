package model

import (
	"gorm.io/gorm"
	"time"
)

type DimensionType string

const (
	App     DimensionType = "app"
	Country DimensionType = "country"
	OS      DimensionType = "os"
)

type TargetingRuleDimension struct {
	ID              string         `json:"id"`
	TargetingRuleID string         `json:"targeting_rule_id"` // FK to TargetingRule
	DimensionType   DimensionType  `json:"dimension_type"`    // "app", "country", "os"
	Include         bool           `json:"include"`           // true = include, false = exclude
	Value           string         `json:"value"`             // actual value, e.g., "android", "us"
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at"`
}

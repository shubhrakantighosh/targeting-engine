package model

import (
	"gorm.io/gorm"
	"time"
)

type TargetingRule struct {
	ID         string         `json:"id"`
	CampaignID string         `json:"campaign_id"`
	Name       string         `json:"name"`
	CreatedBy  string         `json:"created_by"`
	UpdatedBy  string         `json:"updated_by"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at"`
}

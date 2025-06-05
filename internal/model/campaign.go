package model

import (
	"gorm.io/gorm"
	"time"
)

type Campaign struct {
	ID        uint64         `json:"id"`
	Name      string         `json:"name"`
	Image     string         `json:"image"`
	CTA       string         `json:"cta"`
	Status    string         `json:"status"`
	CreatedBy string         `json:"created_by"`
	UpdatedBy string         `json:"updated_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

type Campaigns []Campaign

package model

import (
	"gorm.io/gorm"
	"main/util"
	"time"
)

type Status int

const (
	Active Status = iota + 1
	Inactive
)

func (s Status) Is(status Status) bool {
	return s == status
}

type Campaign struct {
	ID        uint64         `json:"id"`
	Name      string         `json:"name"`
	Image     string         `json:"image"`
	CTA       string         `json:"cta"`
	Status    Status         `json:"status"`
	CreatedBy string         `json:"created_by"`
	UpdatedBy string         `json:"updated_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

type Campaigns []Campaign

func (c Campaigns) IsEmpty() bool {
	return c == nil || len(c) == 0
}

func (c Campaigns) GetActiveCampaignIDs() []uint64 {
	ids := make([]uint64, 0)
	if c.IsEmpty() {
		return ids
	}

	for _, v := range c {
		if v.Status.Is(Active) {
			ids = append(ids, v.ID)
		}
	}

	return util.DeduplicateSlice(ids)
}

func (c Campaigns) GetActiveCampaign() Campaigns {
	campaigns := make(Campaigns, 0)
	if c.IsEmpty() {
		return campaigns
	}

	for _, v := range c {
		if v.Status.Is(Active) {
			campaigns = append(campaigns, v)
		}
	}

	return util.DeduplicateSlice(campaigns)
}

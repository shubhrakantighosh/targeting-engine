package constants

import "time"

const (
	RequestID           = "request_id"
	Env                 = "env"
	Consistency         = "consistency"
	EventualConsistency = "eventual"
	StrongConsistency   = "strong"
	App                 = "app"
	Country             = "country"
	OS                  = "os"
	ID                  = "id"
	DimensionType       = "dimension_type"
	CampaignID          = "campaign_id"
	Value               = "value"
	Include             = "include"
	Exclude             = "exclude"
	Status              = "status"
	OneDay              = time.Hour * 24
)

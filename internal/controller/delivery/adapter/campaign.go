package adapter

import (
	"main/internal/controller/delivery/response"
	"main/internal/model"
)

func TransformCampaignsForController(campaigns model.Campaigns) response.Campaigns {
	campaignsResponse := make(response.Campaigns, 0)
	for _, campaign := range campaigns {
		campaignsResponse = append(campaignsResponse, response.Campaign{
			CID:   campaign.Name,
			Image: campaign.Image,
			CTA:   campaign.CTA,
		})
	}

	return campaignsResponse
}

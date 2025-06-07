package response

type Campaign struct {
	CID   string `json:"cid"`
	Image string `json:"image"`
	CTA   string `json:"cta"`
}

type Campaigns []Campaign

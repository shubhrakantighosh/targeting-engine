package service

import (
	"github.com/google/wire"
	campaignRepository "main/internal/campaign/repository"
	campaignService "main/internal/campaign/service"
	targetingRuleRepository "main/internal/targeting_rule/repository"
	targetingRuleServiec "main/internal/targeting_rule/service"
)

var ProviderSet = wire.NewSet(
	NewService,
	campaignService.NewService,
	targetingRuleServiec.NewService,
	campaignRepository.NewRepository,
	targetingRuleRepository.NewRepository,

	// bind each one of the interfaces
	wire.Bind(new(Interface), new(*Service)),
	wire.Bind(new(campaignService.Interface), new(*campaignService.Service)),
	wire.Bind(new(targetingRuleServiec.Interface), new(*targetingRuleServiec.Service)),
	wire.Bind(new(campaignRepository.Interface), new(*campaignRepository.Repository)),
	wire.Bind(new(targetingRuleRepository.Interface), new(*targetingRuleRepository.Repository)),
)

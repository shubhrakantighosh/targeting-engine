package service

import (
	"main/internal/model"
	"sync"
)

type TargetingFilter struct {
	hierarchy         DimensionHierarchy
	parentChildMap    map[uint64]model.TargetingRules
	campaignRuleIndex map[uint64]model.TargetingRules
}

func NewTargetingFilter(rules model.TargetingRules, hierarchy DimensionHierarchy) *TargetingFilter {
	parentChildMap, campaignRuleIndex := Build(rules)
	return &TargetingFilter{
		hierarchy:         hierarchy,
		parentChildMap:    parentChildMap,
		campaignRuleIndex: campaignRuleIndex,
	}
}

func Build(rules model.TargetingRules) (map[uint64]model.TargetingRules, map[uint64]model.TargetingRules) {
	parentChildMap := make(map[uint64]model.TargetingRules)
	campaignRuleIndex := make(map[uint64]model.TargetingRules)

	for _, rule := range rules {
		if rule.ParentID > 0 {
			parentChildMap[rule.ParentID] = append(parentChildMap[rule.ParentID], rule)
		}

		campaignRuleIndex[rule.CampaignID] = append(campaignRuleIndex[rule.CampaignID], rule)
	}

	return parentChildMap, campaignRuleIndex
}

func (tf *TargetingFilter) GetMatchingCampaigns(campaignIDs []uint64, clientReq ClientRequest) []uint64 {
	type filterResult struct {
		CampaignID uint64
		Matched    bool
	}

	var (
		resultChan = make(chan filterResult, len(campaignIDs))
		wg         = new(sync.WaitGroup)
		results    = make([]uint64, 0)
	)

	wg.Add(len(campaignIDs))

	for _, campaignID := range campaignIDs {
		go func(cID uint64) {
			defer wg.Done()
			matched := tf.matchesCampaign(cID, clientReq)
			resultChan <- filterResult{CampaignID: cID, Matched: matched}
		}(campaignID)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for res := range resultChan {
		if res.Matched {
			results = append(results, res.CampaignID)
		}
	}

	return results
}

func (tf *TargetingFilter) matchesCampaign(campaignID uint64, clientReq ClientRequest) bool {
	matchedRules := make(map[model.DimensionType]uint64)
	rules, ok := tf.campaignRuleIndex[campaignID]
	if !ok {
		return false
	}

	campaignRuleMap := rules.GroupByDimensionAndValue()
	for dimension := range model.NonHierarchicalDimensionType {
		val := clientReq[dimension]
		rulesByVal, found := campaignRuleMap[dimension][val]
		if !found {
			return false
		}

		match := false
		for _, rule := range rulesByVal {
			if rule.Include {
				match = true
				matchedRules[rule.DimensionType] = rule.ID
				break
			}
		}

		if !match {
			return false
		}
	}

	if len(tf.hierarchy.Dimensions) > 0 {
		return tf.checkHierarchicalDimensions(clientReq, matchedRules)
	}
	return true
}

func (tf *TargetingFilter) checkHierarchicalDimensions(
	clientReq ClientRequest,
	matchedRules map[model.DimensionType]uint64,
) bool {
	for _, parentDimType := range tf.hierarchy.Dimensions {
		childDimType := tf.hierarchy.ParentMap[parentDimType]

		clientVal := clientReq[childDimType]
		id, ok := matchedRules[parentDimType]
		if !ok {
			return false
		}

		val, ok := tf.parentChildMap[id]
		if !ok {
			return false
		}

		isFound := false
		for _, rule := range val {
			if rule.Include && rule.Value == clientVal {
				matchedRules[rule.DimensionType] = rule.ID
				isFound = true
				break
			}
		}

		if !isFound {
			return false
		}
	}

	return true
}

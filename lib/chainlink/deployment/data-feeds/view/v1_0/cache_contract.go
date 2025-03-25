package v1_0

import (
	"errors"
	"fmt"

	"github.com/smartcontractkit/chainlink/deployment/common/view/types"
	cache "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/data-feeds/generated/data_feeds_cache"
)

type CacheView struct {
	types.ContractMetaData
}

// GenerateDataFeedsCacheView generates a CacheView from a DataFeedsCache contract.
func GenerateDataFeedsCacheView(cache *cache.DataFeedsCache) (CacheView, error) {
	if cache == nil {
		return CacheView{}, errors.New("cannot generate view for nil DataFeedsCache")
	}
	meta, err := types.NewContractMetaData(cache, cache.Address())
	if err != nil {
		return CacheView{}, fmt.Errorf("failed to generate contract metadata for DataFeedsCache: %w", err)
	}

	return CacheView{
		ContractMetaData: meta,
	}, nil
}

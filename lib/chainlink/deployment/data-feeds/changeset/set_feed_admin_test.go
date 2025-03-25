package changeset_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink/deployment/common/proposalutils"
	commonTypes "github.com/smartcontractkit/chainlink/deployment/common/types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	commonChangesets "github.com/smartcontractkit/chainlink/deployment/common/changeset"

	"github.com/smartcontractkit/chainlink/deployment"
	"github.com/smartcontractkit/chainlink/deployment/data-feeds/changeset"
	"github.com/smartcontractkit/chainlink/deployment/data-feeds/changeset/types"
	"github.com/smartcontractkit/chainlink/deployment/environment/memory"
)

func TestSetCacheAdmin(t *testing.T) {
	t.Parallel()
	lggr := logger.Test(t)
	cfg := memory.MemoryEnvironmentConfig{
		Nodes:  1,
		Chains: 1,
	}
	env := memory.NewMemoryEnvironment(t, lggr, zapcore.DebugLevel, cfg)

	chainSelector := env.AllChainSelectors()[0]

	newEnv, err := commonChangesets.Apply(t, env, nil,
		commonChangesets.Configure(
			changeset.DeployCacheChangeset,
			types.DeployConfig{
				ChainsToDeploy: []uint64{chainSelector},
				Labels:         []string{"data-feeds"},
			},
		),
		commonChangesets.Configure(
			deployment.CreateLegacyChangeSet(commonChangesets.DeployMCMSWithTimelockV2),
			map[uint64]commonTypes.MCMSWithTimelockConfigV2{
				chainSelector: proposalutils.SingleGroupTimelockConfigV2(t),
			},
		),
	)
	require.NoError(t, err)

	cacheAddress, err := deployment.SearchAddressBook(newEnv.ExistingAddresses, chainSelector, "DataFeedsCache")
	require.NoError(t, err)

	// without MCMS
	resp, err := commonChangesets.Apply(t, newEnv, nil,
		commonChangesets.Configure(
			changeset.SetFeedAdminChangeset,
			types.SetFeedAdminConfig{
				ChainSelector: chainSelector,
				CacheAddress:  common.HexToAddress(cacheAddress),
				AdminAddress:  common.HexToAddress("0x123"),
				IsAdmin:       true,
			},
		),
	)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// with MCMS
	newEnv, err = commonChangesets.Apply(t, newEnv, nil,
		commonChangesets.Configure(
			deployment.CreateLegacyChangeSet(commonChangesets.TransferToMCMSWithTimelockV2),
			commonChangesets.TransferToMCMSWithTimelockConfig{
				ContractsByChain: map[uint64][]common.Address{
					chainSelector: {common.HexToAddress(cacheAddress)},
				},
				MinDelay: 0,
			},
		),
	)
	require.NoError(t, err)

	resp, err = commonChangesets.Apply(t, newEnv, nil,
		commonChangesets.Configure(
			changeset.SetFeedAdminChangeset,
			types.SetFeedAdminConfig{
				ChainSelector: chainSelector,
				CacheAddress:  common.HexToAddress(cacheAddress),
				AdminAddress:  common.HexToAddress("0x123"),
				IsAdmin:       true,
				McmsConfig: &types.MCMSConfig{
					MinDelay: 0,
				},
			},
		),
	)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

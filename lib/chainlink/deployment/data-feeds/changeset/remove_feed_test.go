package changeset_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	cache "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/data-feeds/generated/data_feeds_cache"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	commonChangesets "github.com/smartcontractkit/chainlink/deployment/common/changeset"
	"github.com/smartcontractkit/chainlink/deployment/common/proposalutils"
	commonTypes "github.com/smartcontractkit/chainlink/deployment/common/types"
	"github.com/smartcontractkit/chainlink/deployment/data-feeds/shared"

	"github.com/smartcontractkit/chainlink/deployment"
	"github.com/smartcontractkit/chainlink/deployment/data-feeds/changeset"
	"github.com/smartcontractkit/chainlink/deployment/data-feeds/changeset/types"
	"github.com/smartcontractkit/chainlink/deployment/environment/memory"
)

func TestRemoveFeed(t *testing.T) {
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

	dataid, _ := shared.ConvertHexToBytes16("01bb0467f50003040000000000000000")

	// without MCMS
	newEnv, err = commonChangesets.Apply(t, newEnv, nil,
		// set the feed admin, only admin can perform set/remove operations
		commonChangesets.Configure(
			changeset.SetFeedAdminChangeset,
			types.SetFeedAdminConfig{
				ChainSelector: chainSelector,
				CacheAddress:  common.HexToAddress(cacheAddress),
				AdminAddress:  common.HexToAddress(env.Chains[chainSelector].DeployerKey.From.Hex()),
				IsAdmin:       true,
			},
		),
		// set the feed config
		commonChangesets.Configure(
			changeset.SetFeedConfigChangeset,
			types.SetFeedDecimalConfig{
				ChainSelector: chainSelector,
				CacheAddress:  common.HexToAddress(cacheAddress),
				DataIDs:       [][16]byte{dataid},
				Descriptions:  []string{"test"},
				WorkflowMetadata: []cache.DataFeedsCacheWorkflowMetadata{
					cache.DataFeedsCacheWorkflowMetadata{
						AllowedSender:        common.HexToAddress("0x22"),
						AllowedWorkflowOwner: common.HexToAddress("0x33"),
						AllowedWorkflowName:  shared.HashedWorkflowName("test"),
					},
				},
			},
		),
		// remove the feed config
		commonChangesets.Configure(
			changeset.RemoveFeedChangeset,
			types.RemoveFeedConfig{
				ChainSelector:  chainSelector,
				CacheAddress:   common.HexToAddress(cacheAddress),
				DataIDs:        [][16]byte{dataid},
				ProxyAddresses: []common.Address{common.HexToAddress("0x123")},
			},
		),
	)
	require.NoError(t, err)

	// with MCMS
	timeLockAddress, err := deployment.SearchAddressBook(newEnv.ExistingAddresses, chainSelector, "RBACTimelock")
	require.NoError(t, err)

	newEnv, err = commonChangesets.Apply(t, newEnv, nil,
		// Set the admin to the timelock
		commonChangesets.Configure(
			changeset.SetFeedAdminChangeset,
			types.SetFeedAdminConfig{
				ChainSelector: chainSelector,
				CacheAddress:  common.HexToAddress(cacheAddress),
				AdminAddress:  common.HexToAddress(timeLockAddress),
				IsAdmin:       true,
			},
		),
		// Transfer cache ownership to MCMS
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

	// Set and remove the feed config with MCMS
	newEnv, err = commonChangesets.Apply(t, newEnv, nil,
		commonChangesets.Configure(
			changeset.SetFeedConfigChangeset,
			types.SetFeedDecimalConfig{
				ChainSelector: chainSelector,
				CacheAddress:  common.HexToAddress(cacheAddress),
				DataIDs:       [][16]byte{dataid},
				Descriptions:  []string{"test2"},
				WorkflowMetadata: []cache.DataFeedsCacheWorkflowMetadata{
					cache.DataFeedsCacheWorkflowMetadata{
						AllowedSender:        common.HexToAddress("0x22"),
						AllowedWorkflowOwner: common.HexToAddress("0x33"),
						AllowedWorkflowName:  shared.HashedWorkflowName("test"),
					},
				},
				McmsConfig: &types.MCMSConfig{
					MinDelay: 0,
				},
			},
		),
		commonChangesets.Configure(
			changeset.RemoveFeedChangeset,
			types.RemoveFeedConfig{
				ChainSelector:  chainSelector,
				CacheAddress:   common.HexToAddress(cacheAddress),
				DataIDs:        [][16]byte{dataid},
				ProxyAddresses: []common.Address{common.HexToAddress("0x123")},
				McmsConfig: &types.MCMSConfig{
					MinDelay: 0,
				},
			},
		),
	)
	require.NoError(t, err)
}

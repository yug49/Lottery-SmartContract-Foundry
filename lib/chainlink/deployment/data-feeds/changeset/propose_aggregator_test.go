package changeset_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink/deployment"
	"github.com/smartcontractkit/chainlink/deployment/common/proposalutils"
	commonTypes "github.com/smartcontractkit/chainlink/deployment/common/types"

	commonChangesets "github.com/smartcontractkit/chainlink/deployment/common/changeset"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"

	"github.com/smartcontractkit/chainlink/deployment/data-feeds/changeset"
	"github.com/smartcontractkit/chainlink/deployment/data-feeds/changeset/types"
	"github.com/smartcontractkit/chainlink/deployment/environment/memory"
)

func TestProposeAggregator(t *testing.T) {
	t.Parallel()
	lggr := logger.Test(t)
	cfg := memory.MemoryEnvironmentConfig{
		Nodes:  1,
		Chains: 1,
	}
	env := memory.NewMemoryEnvironment(t, lggr, zapcore.DebugLevel, cfg)

	chainSelector := env.AllChainSelectors()[0]

	// without MCMS
	newEnv, err := commonChangesets.Apply(t, env, nil,
		// Deploy cache and aggregator proxy
		commonChangesets.Configure(
			changeset.DeployCacheChangeset,
			types.DeployConfig{
				ChainsToDeploy: []uint64{chainSelector},
				Labels:         []string{"data-feeds"},
			},
		),
		commonChangesets.Configure(
			changeset.DeployAggregatorProxyChangeset,
			types.DeployAggregatorProxyConfig{
				ChainsToDeploy:   []uint64{chainSelector},
				AccessController: []common.Address{common.HexToAddress("0x")},
			},
		),
	)
	require.NoError(t, err)

	proxyAddress, err := deployment.SearchAddressBook(newEnv.ExistingAddresses, chainSelector, "AggregatorProxy")
	require.NoError(t, err)

	newEnv, err = commonChangesets.Apply(t, newEnv, nil,
		// Propose a new aggregator
		commonChangesets.Configure(
			changeset.ProposeAggregatorChangeset,
			types.ProposeConfirmAggregatorConfig{
				ChainSelector:        chainSelector,
				ProxyAddress:         common.HexToAddress(proxyAddress),
				NewAggregatorAddress: common.HexToAddress("0x123"),
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

	// with MCMS
	newEnv, err = commonChangesets.Apply(t, newEnv, nil,
		// transfer proxy ownership to timelock
		commonChangesets.Configure(
			deployment.CreateLegacyChangeSet(commonChangesets.TransferToMCMSWithTimelockV2),
			commonChangesets.TransferToMCMSWithTimelockConfig{
				ContractsByChain: map[uint64][]common.Address{
					chainSelector: {common.HexToAddress(proxyAddress)},
				},
				MinDelay: 0,
			},
		),
		commonChangesets.Configure(
			changeset.ProposeAggregatorChangeset,
			types.ProposeConfirmAggregatorConfig{
				ChainSelector:        chainSelector,
				ProxyAddress:         common.HexToAddress(proxyAddress),
				NewAggregatorAddress: common.HexToAddress("0x123"),
				McmsConfig: &types.MCMSConfig{
					MinDelay: 0,
				},
			},
		),
	)
	require.NoError(t, err)
}

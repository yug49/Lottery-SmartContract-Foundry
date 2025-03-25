package changeset

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	commonChangesets "github.com/smartcontractkit/chainlink/deployment/common/changeset"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"

	"github.com/smartcontractkit/chainlink/deployment/data-feeds/changeset/types"
	"github.com/smartcontractkit/chainlink/deployment/environment/memory"
)

func TestAggregatorProxy(t *testing.T) {
	t.Parallel()
	lggr := logger.Test(t)
	cfg := memory.MemoryEnvironmentConfig{
		Nodes:  1,
		Chains: 2,
	}
	env := memory.NewMemoryEnvironment(t, lggr, zapcore.DebugLevel, cfg)

	chainSelector := env.AllChainSelectors()[0]

	resp, err := commonChangesets.Apply(t, env, nil,
		commonChangesets.Configure(
			DeployCacheChangeset,
			types.DeployConfig{
				ChainsToDeploy: []uint64{chainSelector},
				Labels:         []string{"data-feeds"},
			},
		),
		commonChangesets.Configure(
			DeployAggregatorProxyChangeset,
			types.DeployAggregatorProxyConfig{
				ChainsToDeploy:   []uint64{chainSelector},
				AccessController: []common.Address{common.HexToAddress("0x")},
			},
		),
	)

	require.NoError(t, err)
	require.NotNil(t, resp)

	addrs, err := resp.ExistingAddresses.AddressesForChain(chainSelector)
	require.NoError(t, err)
	require.Len(t, addrs, 2) // AggregatorProxy and DataFeedsCache
}

package testhelpers

import (
	"math/big"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"

	"github.com/smartcontractkit/chainlink/deployment"
	"github.com/smartcontractkit/chainlink/deployment/ccip/changeset"
	"github.com/smartcontractkit/chainlink/deployment/ccip/changeset/v1_5_1"
	commonchangeset "github.com/smartcontractkit/chainlink/deployment/common/changeset"
	commoncs "github.com/smartcontractkit/chainlink/deployment/common/changeset"
	"github.com/smartcontractkit/chainlink/deployment/common/proposalutils"
	commontypes "github.com/smartcontractkit/chainlink/deployment/common/types"
	"github.com/smartcontractkit/chainlink/deployment/environment/memory"
	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/ccip/generated/v1_5_1/token_pool"
	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/shared/generated/burn_mint_erc677"
)

const (
	LocalTokenDecimals                       = 18
	TestTokenSymbol    changeset.TokenSymbol = "TEST"
)

// CreateSymmetricRateLimits is a utility to quickly create a rate limiter config with equal inbound and outbound values.
func CreateSymmetricRateLimits(rate int64, capacity int64) v1_5_1.RateLimiterConfig {
	return v1_5_1.RateLimiterConfig{
		Inbound: token_pool.RateLimiterConfig{
			IsEnabled: rate != 0 || capacity != 0,
			Rate:      big.NewInt(rate),
			Capacity:  big.NewInt(capacity),
		},
		Outbound: token_pool.RateLimiterConfig{
			IsEnabled: rate != 0 || capacity != 0,
			Rate:      big.NewInt(rate),
			Capacity:  big.NewInt(capacity),
		},
	}
}

// SetupTwoChainEnvironmentWithTokens preps the environment for token pool deployment testing.
func SetupTwoChainEnvironmentWithTokens(
	t *testing.T,
	lggr logger.Logger,
	transferToTimelock bool,
) (deployment.Environment, uint64, uint64, map[uint64]*deployment.ContractDeploy[*burn_mint_erc677.BurnMintERC677], map[uint64]*proposalutils.TimelockExecutionContracts) {
	e := memory.NewMemoryEnvironment(t, lggr, zapcore.InfoLevel, memory.MemoryEnvironmentConfig{
		Chains: 2,
	})
	selectors := e.AllChainSelectors()

	addressBook := deployment.NewMemoryAddressBook()
	prereqCfg := make([]changeset.DeployPrerequisiteConfigPerChain, len(selectors))
	for i, selector := range selectors {
		prereqCfg[i] = changeset.DeployPrerequisiteConfigPerChain{
			ChainSelector: selector,
		}
	}

	mcmsCfg := make(map[uint64]commontypes.MCMSWithTimelockConfigV2)
	for _, selector := range selectors {
		mcmsCfg[selector] = proposalutils.SingleGroupTimelockConfigV2(t)
	}

	// Deploy one burn-mint token per chain to use in the tests
	tokens := make(map[uint64]*deployment.ContractDeploy[*burn_mint_erc677.BurnMintERC677])
	for _, selector := range selectors {
		token, err := deployment.DeployContract(e.Logger, e.Chains[selector], addressBook,
			func(chain deployment.Chain) deployment.ContractDeploy[*burn_mint_erc677.BurnMintERC677] {
				tokenAddress, tx, token, err := burn_mint_erc677.DeployBurnMintERC677(
					e.Chains[selector].DeployerKey,
					e.Chains[selector].Client,
					string(TestTokenSymbol),
					string(TestTokenSymbol),
					LocalTokenDecimals,
					big.NewInt(0).Mul(big.NewInt(1e9), big.NewInt(1e18)),
				)
				return deployment.ContractDeploy[*burn_mint_erc677.BurnMintERC677]{
					Address:  tokenAddress,
					Contract: token,
					Tv:       deployment.NewTypeAndVersion(changeset.BurnMintToken, deployment.Version1_0_0),
					Tx:       tx,
					Err:      err,
				}
			},
		)
		require.NoError(t, err)
		tokens[selector] = token
	}

	// Deploy MCMS setup & prerequisite contracts
	e, err := commoncs.Apply(t, e, nil,
		commoncs.Configure(
			deployment.CreateLegacyChangeSet(changeset.DeployPrerequisitesChangeset),
			changeset.DeployPrerequisiteConfig{Configs: prereqCfg},
		),
		commoncs.Configure(
			deployment.CreateLegacyChangeSet(commoncs.DeployMCMSWithTimelockV2),
			mcmsCfg,
		),
	)
	require.NoError(t, err)

	state, err := changeset.LoadOnchainState(e)
	require.NoError(t, err)

	// We only need the token admin registry to be owned by the timelock in these tests
	timelockOwnedContractsByChain := make(map[uint64][]common.Address)
	for _, selector := range selectors {
		timelockOwnedContractsByChain[selector] = []common.Address{state.Chains[selector].TokenAdminRegistry.Address()}
	}

	// Assemble map of addresses required for Timelock scheduling & execution
	timelockContracts := make(map[uint64]*proposalutils.TimelockExecutionContracts)
	for _, selector := range selectors {
		timelockContracts[selector] = &proposalutils.TimelockExecutionContracts{
			Timelock:  state.Chains[selector].Timelock,
			CallProxy: state.Chains[selector].CallProxy,
		}
	}

	if transferToTimelock {
		// Transfer ownership of token admin registry to the Timelock
		e, err = commoncs.Apply(t, e, timelockContracts,
			commoncs.Configure(
				deployment.CreateLegacyChangeSet(commoncs.TransferToMCMSWithTimelock),
				commoncs.TransferToMCMSWithTimelockConfig{
					ContractsByChain: timelockOwnedContractsByChain,
					MinDelay:         0,
				},
			),
		)
		require.NoError(t, err)
	}

	return e, selectors[0], selectors[1], tokens, timelockContracts
}

// getPoolsOwnedByDeployer returns any pools that need to be transferred to timelock.
func getPoolsOwnedByDeployer[T commonchangeset.Ownable](t *testing.T, contracts map[semver.Version]T, chain deployment.Chain) []common.Address {
	var addresses []common.Address
	for _, contract := range contracts {
		owner, err := contract.Owner(nil)
		require.NoError(t, err)
		if owner == chain.DeployerKey.From {
			addresses = append(addresses, contract.Address())
		}
	}
	return addresses
}

// DeployTestTokenPools deploys token pools tied for the TEST token across multiple chains.
func DeployTestTokenPools(
	t *testing.T,
	e deployment.Environment,
	newPools map[uint64]v1_5_1.DeployTokenPoolInput,
	transferToTimelock bool,
) deployment.Environment {
	selectors := e.AllChainSelectors()

	e, err := commonchangeset.Apply(t, e, nil,
		commoncs.Configure(
			deployment.CreateLegacyChangeSet(v1_5_1.DeployTokenPoolContractsChangeset),
			v1_5_1.DeployTokenPoolContractsConfig{
				TokenSymbol: TestTokenSymbol,
				NewPools:    newPools,
			},
		),
	)
	require.NoError(t, err)

	state, err := changeset.LoadOnchainState(e)
	require.NoError(t, err)

	if transferToTimelock {
		// Assemble map of addresses required for Timelock scheduling & execution
		timelockContracts := make(map[uint64]*proposalutils.TimelockExecutionContracts)
		for _, selector := range selectors {
			timelockContracts[selector] = &proposalutils.TimelockExecutionContracts{
				Timelock:  state.Chains[selector].Timelock,
				CallProxy: state.Chains[selector].CallProxy,
			}
		}

		timelockOwnedContractsByChain := make(map[uint64][]common.Address)
		for _, selector := range selectors {
			if newPool, ok := newPools[selector]; ok {
				switch newPool.Type {
				case changeset.BurnFromMintTokenPool:
					timelockOwnedContractsByChain[selector] = getPoolsOwnedByDeployer(t, state.Chains[selector].BurnFromMintTokenPools[TestTokenSymbol], e.Chains[selector])
				case changeset.BurnWithFromMintTokenPool:
					timelockOwnedContractsByChain[selector] = getPoolsOwnedByDeployer(t, state.Chains[selector].BurnWithFromMintTokenPools[TestTokenSymbol], e.Chains[selector])
				case changeset.BurnMintTokenPool:
					timelockOwnedContractsByChain[selector] = getPoolsOwnedByDeployer(t, state.Chains[selector].BurnMintTokenPools[TestTokenSymbol], e.Chains[selector])
				case changeset.LockReleaseTokenPool:
					timelockOwnedContractsByChain[selector] = getPoolsOwnedByDeployer(t, state.Chains[selector].LockReleaseTokenPools[TestTokenSymbol], e.Chains[selector])
				}
			}
		}

		// Transfer ownership of token admin registry to the Timelock
		e, err = commoncs.Apply(t, e, timelockContracts,
			commoncs.Configure(
				deployment.CreateLegacyChangeSet(commoncs.TransferToMCMSWithTimelock),
				commoncs.TransferToMCMSWithTimelockConfig{
					ContractsByChain: timelockOwnedContractsByChain,
					MinDelay:         0,
				},
			),
		)
		require.NoError(t, err)
	}

	return e
}

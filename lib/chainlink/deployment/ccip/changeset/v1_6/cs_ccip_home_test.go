package v1_6_test

import (
	"math/big"
	"regexp"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"

	"github.com/smartcontractkit/chainlink-common/pkg/config"

	"github.com/smartcontractkit/chainlink-ccip/chainconfig"
	cciptypes "github.com/smartcontractkit/chainlink-ccip/pkg/types/ccipocr3"
	"github.com/smartcontractkit/chainlink-testing-framework/lib/utils/testcontext"

	"github.com/smartcontractkit/chainlink/deployment"
	"github.com/smartcontractkit/chainlink/deployment/ccip/changeset"
	"github.com/smartcontractkit/chainlink/deployment/ccip/changeset/globals"
	"github.com/smartcontractkit/chainlink/deployment/ccip/changeset/internal"
	"github.com/smartcontractkit/chainlink/deployment/ccip/changeset/testhelpers"
	"github.com/smartcontractkit/chainlink/deployment/ccip/changeset/v1_6"
	"github.com/smartcontractkit/chainlink/v2/core/capabilities/ccip/types"
	"github.com/smartcontractkit/chainlink/v2/core/logger"

	"github.com/stretchr/testify/require"

	commonchangeset "github.com/smartcontractkit/chainlink/deployment/common/changeset"
	"github.com/smartcontractkit/chainlink/deployment/common/proposalutils"
)

func TestInvalidOCR3Params(t *testing.T) {
	e, _ := testhelpers.NewMemoryEnvironment(t,
		testhelpers.WithPrerequisiteDeploymentOnly(nil))
	chain1 := e.Env.AllChainSelectors()[0]
	envNodes, err := deployment.NodeInfo(e.Env.NodeIDs, e.Env.Offchain)
	require.NoError(t, err)
	// Need to deploy prerequisites first so that we can form the USDC config
	// no proposals to be made, timelock can be passed as nil here
	e.Env, err = commonchangeset.Apply(t, e.Env, nil,
		commonchangeset.Configure(
			deployment.CreateLegacyChangeSet(v1_6.DeployHomeChainChangeset),
			v1_6.DeployHomeChainConfig{
				HomeChainSel:     e.HomeChainSel,
				RMNDynamicConfig: testhelpers.NewTestRMNDynamicConfig(),
				RMNStaticConfig:  testhelpers.NewTestRMNStaticConfig(),
				NodeOperators:    testhelpers.NewTestNodeOperator(e.Env.Chains[e.HomeChainSel].DeployerKey.From),
				NodeP2PIDsPerNodeOpAdmin: map[string][][32]byte{
					testhelpers.TestNodeOperator: envNodes.NonBootstraps().PeerIDs(),
				},
			},
		),
		commonchangeset.Configure(
			deployment.CreateLegacyChangeSet(v1_6.DeployChainContractsChangeset),
			v1_6.DeployChainContractsConfig{
				HomeChainSelector: e.HomeChainSel,
				ContractParamsPerChain: map[uint64]v1_6.ChainContractParams{
					chain1: {
						FeeQuoterParams: v1_6.DefaultFeeQuoterParams(),
						OffRampParams:   v1_6.DefaultOffRampParams(),
					},
				},
			},
		),
	)
	require.NoError(t, err)

	state, err := changeset.LoadOnchainState(e.Env)
	require.NoError(t, err)
	nodes, err := deployment.NodeInfo(e.Env.NodeIDs, e.Env.Offchain)
	require.NoError(t, err)
	params := v1_6.DeriveOCRParamsForCommit(v1_6.SimulationTest, e.FeedChainSel, nil, nil)

	// tweak params to have invalid config
	// make DeltaRound greater than DeltaProgress
	params.OCRParameters.DeltaRound = params.OCRParameters.DeltaProgress + time.Duration(1)
	_, err = internal.BuildOCR3ConfigForCCIPHome(
		state.Chains[e.HomeChainSel].CCIPHome,
		e.Env.OCRSecrets,
		state.Chains[chain1].OffRamp.Address().Bytes(),
		chain1,
		nodes.NonBootstraps(),
		state.Chains[e.HomeChainSel].RMNHome.Address(),
		params.OCRParameters,
		params.CommitOffChainConfig,
		&globals.DefaultExecuteOffChainCfg,
		false,
	)
	require.Errorf(t, err, "expected error")
	pattern := `DeltaRound \(\d+\.\d+s\) must be less than DeltaProgress \(\d+s\)`
	matched, err1 := regexp.MatchString(pattern, err.Error())
	require.NoError(t, err1)
	require.True(t, matched)
}

func Test_PromoteCandidate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		mcmsEnabled bool
	}{
		{
			name:        "MCMS enabled",
			mcmsEnabled: true,
		},
		{
			name:        "MCMS disabled",
			mcmsEnabled: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := testcontext.Get(t)
			tenv, _ := testhelpers.NewMemoryEnvironment(t,
				testhelpers.WithNumOfChains(2),
				testhelpers.WithNumOfNodes(4))
			state, err := changeset.LoadOnchainState(tenv.Env)
			require.NoError(t, err)

			// Deploy to all chains.
			allChains := maps.Keys(tenv.Env.Chains)
			source := allChains[0]
			dest := allChains[1]

			if tc.mcmsEnabled {
				// Transfer ownership to timelock so that we can promote the zero digest later down the line.
				transferToTimelock(t, tenv, state, source, dest)
			}

			var (
				capReg   = state.Chains[tenv.HomeChainSel].CapabilityRegistry
				ccipHome = state.Chains[tenv.HomeChainSel].CCIPHome
			)
			donID, err := internal.DonIDForChain(capReg, ccipHome, dest)
			require.NoError(t, err)
			require.NotEqual(t, uint32(0), donID)
			t.Logf("donID: %d", donID)
			candidateDigestCommitBefore, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPCommit))
			require.NoError(t, err)
			require.Equal(t, [32]byte{}, candidateDigestCommitBefore)
			ActiveDigestExecBefore, err := ccipHome.GetActiveDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPExec))
			require.NoError(t, err)
			require.NotEqual(t, [32]byte{}, ActiveDigestExecBefore)

			var mcmsConfig *changeset.MCMSConfig
			if tc.mcmsEnabled {
				mcmsConfig = &changeset.MCMSConfig{
					MinDelay: 0,
				}
			}
			// promotes zero digest on commit and ensure exec is not affected
			_, err = commonchangeset.Apply(t, tenv.Env,
				map[uint64]*proposalutils.TimelockExecutionContracts{
					tenv.HomeChainSel: {
						Timelock:  state.Chains[tenv.HomeChainSel].Timelock,
						CallProxy: state.Chains[tenv.HomeChainSel].CallProxy,
					},
				},
				commonchangeset.Configure(
					deployment.CreateLegacyChangeSet(v1_6.PromoteCandidateChangeset),
					v1_6.PromoteCandidateChangesetConfig{
						HomeChainSelector: tenv.HomeChainSel,
						PluginInfo: []v1_6.PromoteCandidatePluginInfo{
							{
								RemoteChainSelectors:    []uint64{dest},
								PluginType:              types.PluginTypeCCIPCommit,
								AllowEmptyConfigPromote: true,
							},
						},
						MCMS: mcmsConfig,
					},
				),
			)
			require.NoError(t, err)

			// after promoting the zero digest, active digest should also be zero
			activeDigestCommit, err := ccipHome.GetActiveDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPCommit))
			require.NoError(t, err)
			require.Equal(t, [32]byte{}, activeDigestCommit)

			activeDigestExec, err := ccipHome.GetActiveDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPExec))
			require.NoError(t, err)
			require.Equal(t, ActiveDigestExecBefore, activeDigestExec)
		})
	}
}

func Test_SetCandidate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		mcmsEnabled bool
	}{
		{
			name:        "MCMS enabled",
			mcmsEnabled: true,
		},
		{
			name:        "MCMS disabled",
			mcmsEnabled: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := testcontext.Get(t)
			tenv, _ := testhelpers.NewMemoryEnvironment(t,
				testhelpers.WithNumOfChains(2),
				testhelpers.WithNumOfNodes(4))
			state, err := changeset.LoadOnchainState(tenv.Env)
			require.NoError(t, err)

			// Deploy to all chains.
			allChains := maps.Keys(tenv.Env.Chains)
			source := allChains[0]
			dest := allChains[1]

			if tc.mcmsEnabled {
				// Transfer ownership to timelock so that we can promote the zero digest later down the line.
				transferToTimelock(t, tenv, state, source, dest)
			}

			var (
				capReg   = state.Chains[tenv.HomeChainSel].CapabilityRegistry
				ccipHome = state.Chains[tenv.HomeChainSel].CCIPHome
			)
			donID, err := internal.DonIDForChain(capReg, ccipHome, dest)
			require.NoError(t, err)
			require.NotEqual(t, uint32(0), donID)
			candidateDigestCommitBefore, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPCommit))
			require.NoError(t, err)
			require.Equal(t, [32]byte{}, candidateDigestCommitBefore)
			candidateDigestExecBefore, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPExec))
			require.NoError(t, err)
			require.Equal(t, [32]byte{}, candidateDigestExecBefore)

			var mcmsConfig *changeset.MCMSConfig
			if tc.mcmsEnabled {
				mcmsConfig = &changeset.MCMSConfig{
					MinDelay: 0,
				}
			}

			tokenConfig := changeset.NewTestTokenConfig(state.Chains[tenv.FeedChainSel].USDFeeds)
			_, err = commonchangeset.Apply(t, tenv.Env,
				map[uint64]*proposalutils.TimelockExecutionContracts{
					tenv.HomeChainSel: {
						Timelock:  state.Chains[tenv.HomeChainSel].Timelock,
						CallProxy: state.Chains[tenv.HomeChainSel].CallProxy,
					},
				},
				commonchangeset.Configure(
					deployment.CreateLegacyChangeSet(v1_6.SetCandidateChangeset),
					v1_6.SetCandidateChangesetConfig{
						SetCandidateConfigBase: v1_6.SetCandidateConfigBase{
							HomeChainSelector: tenv.HomeChainSel,
							FeedChainSelector: tenv.FeedChainSel,
							MCMS:              mcmsConfig,
						},
						PluginInfo: []v1_6.SetCandidatePluginInfo{
							{
								OCRConfigPerRemoteChainSelector: map[uint64]v1_6.CCIPOCRParams{
									dest: v1_6.DeriveOCRParamsForCommit(v1_6.SimulationTest, tenv.FeedChainSel, tokenConfig.GetTokenInfo(logger.TestLogger(t),
										state.Chains[dest].LinkToken.Address(),
										state.Chains[dest].Weth9.Address()), nil),
								},
								PluginType: types.PluginTypeCCIPCommit,
							},
							{
								OCRConfigPerRemoteChainSelector: map[uint64]v1_6.CCIPOCRParams{
									dest: v1_6.DeriveOCRParamsForExec(v1_6.SimulationTest, nil, func(params v1_6.CCIPOCRParams) v1_6.CCIPOCRParams {
										dCfg, err := state.Chains[dest].OffRamp.GetDynamicConfig(&bind.CallOpts{
											Context: ctx,
										})
										require.NoError(t, err)
										params.ExecuteOffChainConfig.MessageVisibilityInterval =
											*config.MustNewDuration(
												time.Duration(dCfg.PermissionLessExecutionThresholdSeconds + uint32(time.Second)))
										return params
									}),
								},
								PluginType: types.PluginTypeCCIPExec,
							},
						},
					},
				),
			)
			require.Error(t, err)
			require.Contains(t, err.Error(), "does not match the permissionlessExecutionThresholdSeconds in dynamic config")

			// now set the correct config
			_, err = commonchangeset.Apply(t, tenv.Env,
				map[uint64]*proposalutils.TimelockExecutionContracts{
					tenv.HomeChainSel: {
						Timelock:  state.Chains[tenv.HomeChainSel].Timelock,
						CallProxy: state.Chains[tenv.HomeChainSel].CallProxy,
					},
				},
				commonchangeset.Configure(
					deployment.CreateLegacyChangeSet(v1_6.SetCandidateChangeset),
					v1_6.SetCandidateChangesetConfig{
						SetCandidateConfigBase: v1_6.SetCandidateConfigBase{
							HomeChainSelector: tenv.HomeChainSel,
							FeedChainSelector: tenv.FeedChainSel,
							MCMS:              mcmsConfig,
						},
						PluginInfo: []v1_6.SetCandidatePluginInfo{
							{
								OCRConfigPerRemoteChainSelector: map[uint64]v1_6.CCIPOCRParams{
									dest: v1_6.DeriveOCRParamsForCommit(v1_6.SimulationTest, tenv.FeedChainSel, tokenConfig.GetTokenInfo(logger.TestLogger(t),
										state.Chains[dest].LinkToken.Address(),
										state.Chains[dest].Weth9.Address()), nil),
								},
								PluginType: types.PluginTypeCCIPCommit,
							},
							{
								OCRConfigPerRemoteChainSelector: map[uint64]v1_6.CCIPOCRParams{
									dest: v1_6.DeriveOCRParamsForExec(v1_6.SimulationTest, nil, nil),
								},
								PluginType: types.PluginTypeCCIPExec,
							},
						},
					},
				),
			)
			require.NoError(t, err)
			// after setting a new candidate on both plugins, the candidate config digest
			// should be nonzero.
			candidateDigestCommitAfter, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPCommit))
			require.NoError(t, err)
			require.NotEqual(t, [32]byte{}, candidateDigestCommitAfter)
			require.NotEqual(t, candidateDigestCommitBefore, candidateDigestCommitAfter)

			candidateDigestExecAfter, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPExec))
			require.NoError(t, err)
			require.NotEqual(t, [32]byte{}, candidateDigestExecAfter)
			require.NotEqual(t, candidateDigestExecBefore, candidateDigestExecAfter)
		})
	}
}

func Test_RevokeCandidate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		mcmsEnabled bool
	}{
		{
			name:        "MCMS enabled",
			mcmsEnabled: true,
		},
		{
			name:        "MCMS disabled",
			mcmsEnabled: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := testcontext.Get(t)
			tenv, _ := testhelpers.NewMemoryEnvironment(t,
				testhelpers.WithNumOfChains(2),
				testhelpers.WithNumOfNodes(4))
			state, err := changeset.LoadOnchainState(tenv.Env)
			require.NoError(t, err)

			// Deploy to all chains.
			allChains := maps.Keys(tenv.Env.Chains)
			source := allChains[0]
			dest := allChains[1]

			if tc.mcmsEnabled {
				// Transfer ownership to timelock so that we can promote the zero digest later down the line.
				transferToTimelock(t, tenv, state, source, dest)
			}

			var (
				capReg   = state.Chains[tenv.HomeChainSel].CapabilityRegistry
				ccipHome = state.Chains[tenv.HomeChainSel].CCIPHome
			)
			donID, err := internal.DonIDForChain(capReg, ccipHome, dest)
			require.NoError(t, err)
			require.NotEqual(t, uint32(0), donID)
			candidateDigestCommitBefore, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPCommit))
			require.NoError(t, err)
			require.Equal(t, [32]byte{}, candidateDigestCommitBefore)
			candidateDigestExecBefore, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPExec))
			require.NoError(t, err)
			require.Equal(t, [32]byte{}, candidateDigestExecBefore)

			var mcmsConfig *changeset.MCMSConfig
			if tc.mcmsEnabled {
				mcmsConfig = &changeset.MCMSConfig{
					MinDelay: 0,
				}
			}
			tokenConfig := changeset.NewTestTokenConfig(state.Chains[tenv.FeedChainSel].USDFeeds)
			_, err = commonchangeset.Apply(t, tenv.Env,
				map[uint64]*proposalutils.TimelockExecutionContracts{
					tenv.HomeChainSel: {
						Timelock:  state.Chains[tenv.HomeChainSel].Timelock,
						CallProxy: state.Chains[tenv.HomeChainSel].CallProxy,
					},
				},
				commonchangeset.Configure(
					deployment.CreateLegacyChangeSet(v1_6.SetCandidateChangeset),
					v1_6.SetCandidateChangesetConfig{
						SetCandidateConfigBase: v1_6.SetCandidateConfigBase{
							HomeChainSelector: tenv.HomeChainSel,
							FeedChainSelector: tenv.FeedChainSel,
							MCMS:              mcmsConfig,
						},
						PluginInfo: []v1_6.SetCandidatePluginInfo{
							{
								OCRConfigPerRemoteChainSelector: map[uint64]v1_6.CCIPOCRParams{
									dest: v1_6.DeriveOCRParamsForCommit(v1_6.SimulationTest, tenv.FeedChainSel, tokenConfig.GetTokenInfo(logger.TestLogger(t),
										state.Chains[dest].LinkToken.Address(),
										state.Chains[dest].Weth9.Address()), nil),
								},
								PluginType: types.PluginTypeCCIPCommit,
							},
							{
								OCRConfigPerRemoteChainSelector: map[uint64]v1_6.CCIPOCRParams{
									dest: v1_6.DeriveOCRParamsForExec(v1_6.SimulationTest, nil, nil),
								},
								PluginType: types.PluginTypeCCIPExec,
							},
						},
					},
				),
			)
			require.NoError(t, err)

			// after setting a new candidate on both plugins, the candidate config digest
			// should be nonzero.
			candidateDigestCommitAfter, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPCommit))
			require.NoError(t, err)
			require.NotEqual(t, [32]byte{}, candidateDigestCommitAfter)
			require.NotEqual(t, candidateDigestCommitBefore, candidateDigestCommitAfter)

			candidateDigestExecAfter, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPExec))
			require.NoError(t, err)
			require.NotEqual(t, [32]byte{}, candidateDigestExecAfter)
			require.NotEqual(t, candidateDigestExecBefore, candidateDigestExecAfter)

			// next we can revoke candidate - this should set the candidate digest back to zero
			_, err = commonchangeset.Apply(t, tenv.Env,
				map[uint64]*proposalutils.TimelockExecutionContracts{
					tenv.HomeChainSel: {
						Timelock:  state.Chains[tenv.HomeChainSel].Timelock,
						CallProxy: state.Chains[tenv.HomeChainSel].CallProxy,
					},
				},
				commonchangeset.Configure(
					deployment.CreateLegacyChangeSet(v1_6.RevokeCandidateChangeset),
					v1_6.RevokeCandidateChangesetConfig{
						HomeChainSelector:   tenv.HomeChainSel,
						RemoteChainSelector: dest,
						PluginType:          types.PluginTypeCCIPCommit,
						MCMS:                mcmsConfig,
					},
				),
				commonchangeset.Configure(
					deployment.CreateLegacyChangeSet(v1_6.RevokeCandidateChangeset),
					v1_6.RevokeCandidateChangesetConfig{
						HomeChainSelector:   tenv.HomeChainSel,
						RemoteChainSelector: dest,
						PluginType:          types.PluginTypeCCIPExec,
						MCMS:                mcmsConfig,
					},
				),
			)
			require.NoError(t, err)

			// after revoking the candidate, the candidate digest should be zero
			candidateDigestCommitAfterRevoke, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPCommit))
			require.NoError(t, err)
			require.Equal(t, [32]byte{}, candidateDigestCommitAfterRevoke)

			candidateDigestExecAfterRevoke, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPExec))
			require.NoError(t, err)
			require.Equal(t, [32]byte{}, candidateDigestExecAfterRevoke)
		})
	}
}

func transferToTimelock(
	t *testing.T,
	tenv testhelpers.DeployedEnv,
	state changeset.CCIPOnChainState,
	source,
	dest uint64) {
	// Transfer ownership to timelock so that we can promote the zero digest later down the line.
	_, err := commonchangeset.Apply(t, tenv.Env,
		map[uint64]*proposalutils.TimelockExecutionContracts{
			source: {
				Timelock:  state.Chains[source].Timelock,
				CallProxy: state.Chains[source].CallProxy,
			},
			dest: {
				Timelock:  state.Chains[dest].Timelock,
				CallProxy: state.Chains[dest].CallProxy,
			},
			tenv.HomeChainSel: {
				Timelock:  state.Chains[tenv.HomeChainSel].Timelock,
				CallProxy: state.Chains[tenv.HomeChainSel].CallProxy,
			},
		},
		commonchangeset.Configure(
			deployment.CreateLegacyChangeSet(commonchangeset.TransferToMCMSWithTimelock),
			testhelpers.GenTestTransferOwnershipConfig(tenv, []uint64{source, dest}, state),
		),
	)
	require.NoError(t, err)
	testhelpers.AssertTimelockOwnership(t, tenv, []uint64{source, dest}, state)
}

func Test_UpdateChainConfigs(t *testing.T) {
	for _, tc := range []struct {
		name        string
		mcmsEnabled bool
	}{
		{
			name:        "MCMS enabled",
			mcmsEnabled: true,
		},
		{
			name:        "MCMS disabled",
			mcmsEnabled: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tenv, _ := testhelpers.NewMemoryEnvironment(t, testhelpers.WithNumOfChains(3))
			state, err := changeset.LoadOnchainState(tenv.Env)
			require.NoError(t, err)

			allChains := maps.Keys(tenv.Env.Chains)
			source := allChains[0]
			dest := allChains[1]
			otherChain := allChains[2]

			if tc.mcmsEnabled {
				// Transfer ownership to timelock so that we can promote the zero digest later down the line.
				transferToTimelock(t, tenv, state, source, dest)
			}

			ccipHome := state.Chains[tenv.HomeChainSel].CCIPHome
			otherChainConfig, err := ccipHome.GetChainConfig(nil, otherChain)
			require.NoError(t, err)
			assert.NotZero(t, otherChainConfig.FChain)

			var mcmsConfig *changeset.MCMSConfig
			if tc.mcmsEnabled {
				mcmsConfig = &changeset.MCMSConfig{
					MinDelay: 0,
				}
			}
			_, err = commonchangeset.Apply(t, tenv.Env,
				map[uint64]*proposalutils.TimelockExecutionContracts{
					tenv.HomeChainSel: {
						Timelock:  state.Chains[tenv.HomeChainSel].Timelock,
						CallProxy: state.Chains[tenv.HomeChainSel].CallProxy,
					},
				},
				commonchangeset.Configure(
					deployment.CreateLegacyChangeSet(v1_6.UpdateChainConfigChangeset),
					v1_6.UpdateChainConfigConfig{
						HomeChainSelector:  tenv.HomeChainSel,
						RemoteChainRemoves: []uint64{otherChain},
						RemoteChainAdds:    make(map[uint64]v1_6.ChainConfig),
						MCMS:               mcmsConfig,
					},
				),
			)
			require.NoError(t, err)

			// other chain should be gone
			chainConfigAfter, err := ccipHome.GetChainConfig(nil, otherChain)
			require.NoError(t, err)
			assert.Zero(t, chainConfigAfter.FChain)

			// Lets add it back now.
			_, err = commonchangeset.Apply(t, tenv.Env,
				map[uint64]*proposalutils.TimelockExecutionContracts{
					tenv.HomeChainSel: {
						Timelock:  state.Chains[tenv.HomeChainSel].Timelock,
						CallProxy: state.Chains[tenv.HomeChainSel].CallProxy,
					},
				},
				commonchangeset.Configure(
					deployment.CreateLegacyChangeSet(v1_6.UpdateChainConfigChangeset),
					v1_6.UpdateChainConfigConfig{
						HomeChainSelector:  tenv.HomeChainSel,
						RemoteChainRemoves: []uint64{},
						RemoteChainAdds: map[uint64]v1_6.ChainConfig{
							otherChain: {
								EncodableChainConfig: chainconfig.ChainConfig{
									GasPriceDeviationPPB:    cciptypes.BigInt{Int: big.NewInt(globals.GasPriceDeviationPPB)},
									DAGasPriceDeviationPPB:  cciptypes.BigInt{Int: big.NewInt(globals.DAGasPriceDeviationPPB)},
									OptimisticConfirmations: globals.OptimisticConfirmations,
								},
								FChain:  otherChainConfig.FChain,
								Readers: otherChainConfig.Readers,
							},
						},
						MCMS: mcmsConfig,
					},
				),
			)
			require.NoError(t, err)

			chainConfigAfter2, err := ccipHome.GetChainConfig(nil, otherChain)
			require.NoError(t, err)
			assert.Equal(t, chainConfigAfter2.FChain, otherChainConfig.FChain)
			assert.Equal(t, chainConfigAfter2.Readers, otherChainConfig.Readers)
			assert.Equal(t, chainConfigAfter2.Config, otherChainConfig.Config)
		})
	}
}

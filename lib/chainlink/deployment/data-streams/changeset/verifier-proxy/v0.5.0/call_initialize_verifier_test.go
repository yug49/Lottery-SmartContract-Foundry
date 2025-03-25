package v0_5_0

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	verifier_v0_5_0 "github.com/smartcontractkit/chainlink/deployment/data-streams/changeset/verifier/v0_5_0"

	"github.com/smartcontractkit/chainlink/deployment"
	commonChangesets "github.com/smartcontractkit/chainlink/deployment/common/changeset"
	"github.com/smartcontractkit/chainlink/deployment/data-streams/changeset/testutil"
	"github.com/smartcontractkit/chainlink/deployment/data-streams/changeset/types"
	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/llo-feeds/generated/verifier_proxy_v0_5_0"
)

func TestInitializeVerifier(t *testing.T) {
	e := testutil.NewMemoryEnv(t, true, 0)

	chainSelector := e.AllChainSelectors()[0]
	e, err := commonChangesets.Apply(t, e, nil,
		commonChangesets.Configure(
			DeployVerifierProxyChangeset,
			DeployVerifierProxyConfig{
				ChainsToDeploy: []uint64{chainSelector},
				Version:        *semver.MustParse("0.5.0"),
			},
		),
	)
	require.NoError(t, err)

	// Ensure the VerifierProxy was deployed
	verifierProxyAddrHex, err := deployment.SearchAddressBook(e.ExistingAddresses, chainSelector, types.VerifierProxy)
	require.NoError(t, err)
	verifierProxyAddr := common.HexToAddress(verifierProxyAddrHex)

	// Deploy Verifier
	e, err = commonChangesets.Apply(t, e, nil,
		commonChangesets.Configure(
			verifier_v0_5_0.DeployVerifierChangeset,
			verifier_v0_5_0.DeployVerifierConfig{
				ChainsToDeploy: map[uint64]verifier_v0_5_0.DeployVerifier{
					chainSelector: {
						VerifierProxyAddress: verifierProxyAddr,
					},
				},
			},
		),
	)
	require.NoError(t, err)

	// Ensure the Verifier was deployed
	verifierAddrHex, err := deployment.SearchAddressBook(e.ExistingAddresses, chainSelector, types.Verifier)
	require.NoError(t, err)
	verifierAddr := common.HexToAddress(verifierAddrHex)

	e, err = commonChangesets.Apply(t, e, nil,
		commonChangesets.Configure(
			InitializeVerifierChangeset,
			VerifierProxyInitializeVerifierConfig{
				ConfigPerChain: map[uint64][]InitializeVerifierConfig{
					chainSelector: {
						{VerifierAddress: verifierAddr, ContractAddress: verifierProxyAddr},
					},
				},
			},
		),
	)
	require.NoError(t, err)

	chain := e.Chains[chainSelector]

	vp, err := verifier_proxy_v0_5_0.NewVerifierProxy(verifierProxyAddr, chain.Client)
	require.NoError(t, err)
	logIterator, err := vp.FilterVerifierInitialized(nil)
	require.NoError(t, err)

	foundExpected := false
	for logIterator.Next() {
		if verifierAddr == logIterator.Event.VerifierAddress {
			foundExpected = true
			break
		}
	}
	require.True(t, foundExpected)
	err = logIterator.Close()
	if err != nil {
		t.Errorf("Error closing log iterator: %v", err)
	}
}

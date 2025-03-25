package v1_6_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-testing-framework/lib/utils/testcontext"

	"github.com/smartcontractkit/chainlink/deployment/ccip/changeset"
	"github.com/smartcontractkit/chainlink/deployment/ccip/changeset/testhelpers"
	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/ccip/generated/v1_2_0/router"
)

func TestAddLanesWithTestRouter(t *testing.T) {
	t.Parallel()
	e, _ := testhelpers.NewMemoryEnvironment(t)
	// Here we have CR + nodes set up, but no CCIP contracts deployed.
	state, err := changeset.LoadOnchainState(e.Env)
	require.NoError(t, err)

	selectors := e.Env.AllChainSelectors()
	chain1, chain2 := selectors[0], selectors[1]
	testhelpers.AddLaneWithDefaultPricesAndFeeQuoterConfig(t, &e, state, chain1, chain2, true)
	// Need to keep track of the block number for each chain so that event subscription can be done from that block.
	startBlocks := make(map[uint64]*uint64)
	// Send a message from each chain to every other chain.
	expectedSeqNumExec := make(map[testhelpers.SourceDestPair][]uint64)
	latesthdr, err := e.Env.Chains[chain2].Client.HeaderByNumber(testcontext.Get(t), nil)
	require.NoError(t, err)
	block := latesthdr.Number.Uint64()
	startBlocks[chain2] = &block
	msgSentEvent := testhelpers.TestSendRequest(t, e.Env, state, chain1, chain2, true, router.ClientEVM2AnyMessage{
		Receiver:     common.LeftPadBytes(state.Chains[chain2].Receiver.Address().Bytes(), 32),
		Data:         []byte("hello"),
		TokenAmounts: nil,
		FeeToken:     common.HexToAddress("0x0"),
		ExtraArgs:    nil,
	})
	expectedSeqNumExec[testhelpers.SourceDestPair{
		SourceChainSelector: chain1,
		DestChainSelector:   chain2,
	}] = []uint64{msgSentEvent.SequenceNumber}
	testhelpers.ConfirmExecWithSeqNrsForAll(t, e.Env, state, expectedSeqNumExec, startBlocks)
}

// dev is on going for sending request between solana and evm chains
// this test is there to ensure addLane works between solana and evm chains
func TestAddLanesWithSolana(t *testing.T) {
	t.Parallel()
	e, _ := testhelpers.NewMemoryEnvironment(t, testhelpers.WithSolChains(1))
	// Here we have CR + nodes set up, but no CCIP contracts deployed.
	state, err := changeset.LoadOnchainState(e.Env)
	require.NoError(t, err)

	evmSelectors := e.Env.AllChainSelectors()
	chain1, chain2 := evmSelectors[0], evmSelectors[1]
	solSelectors := e.Env.AllChainSelectorsSolana()
	solChain := solSelectors[0]
	testhelpers.AddLaneWithDefaultPricesAndFeeQuoterConfig(t, &e, state, chain1, solChain, true)
	// AddLaneWithDefaultPricesAndFeeQuoterConfig involves calling AddRemoteChainToSolana
	// which adds chain1 to solana
	// so we can not call AddRemoteChainToSolana again with chain1 again, hence using chain2 below
	testhelpers.AddLaneWithDefaultPricesAndFeeQuoterConfig(t, &e, state, solChain, chain2, true)
	_, _, _, err = testhelpers.DeployTransferableTokenSolana(t, e.Env.Logger, e.Env, chain1, solChain, e.Env.Chains[chain1].DeployerKey, e.Env.ExistingAddresses, "MY_TOKEN")
	require.NoError(t, err)
}

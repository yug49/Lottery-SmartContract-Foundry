package framework

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/workflow/generated/workflow_registry_wrapper"
)

type WorkflowRegistry struct {
	t        *testing.T
	backend  *EthBlockchain
	contract *workflow_registry_wrapper.WorkflowRegistry
	addr     common.Address
}

func NewWorkflowRegistry(ctx context.Context, t *testing.T, backend *EthBlockchain) *WorkflowRegistry {
	// Deploy a test workflow_registry
	wfRegistryAddr, _, wfRegistryC, err := workflow_registry_wrapper.DeployWorkflowRegistry(backend.transactionOpts, backend.Backend.Client())
	backend.Backend.Commit()
	require.NoError(t, err)

	// setup contract state to allow the secrets to be updated
	updateAuthorizedAddress(t, backend, wfRegistryC, []common.Address{backend.transactionOpts.From}, true)

	return &WorkflowRegistry{t: t, addr: wfRegistryAddr, contract: wfRegistryC, backend: backend}
}

func (r *WorkflowRegistry) UpdateAllowedDons(donIDs []uint32) {
	updateAllowedDONs(r.t, r.backend, r.contract, donIDs, true)
}

func (r *WorkflowRegistry) RegisterWorkflow(input Workflow, donID uint32) {
	registerWorkflow(r.t, r.backend, r.contract, input, donID)
}

func updateAuthorizedAddress(
	t *testing.T,
	th *EthBlockchain,
	wfRegC *workflow_registry_wrapper.WorkflowRegistry,
	addresses []common.Address,
	_ bool,
) {
	t.Helper()
	_, err := wfRegC.UpdateAuthorizedAddresses(th.transactionOpts, addresses, true)
	require.NoError(t, err, "failed to update authorised addresses")
	th.Backend.Commit()
	th.Backend.Commit()
	th.Backend.Commit()
	gotAddresses, err := wfRegC.GetAllAuthorizedAddresses(&bind.CallOpts{
		From: th.transactionOpts.From,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, addresses, gotAddresses)
}

func updateAllowedDONs(
	t *testing.T,
	th *EthBlockchain,
	wfRegC *workflow_registry_wrapper.WorkflowRegistry,
	donIDs []uint32,
	allowed bool,
) {
	t.Helper()
	_, err := wfRegC.UpdateAllowedDONs(th.transactionOpts, donIDs, allowed)
	require.NoError(t, err, "failed to update DONs")
	th.Backend.Commit()
	th.Backend.Commit()
	th.Backend.Commit()
	gotDons, err := wfRegC.GetAllAllowedDONs(&bind.CallOpts{
		From: th.transactionOpts.From,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, donIDs, gotDons)
}

type Workflow struct {
	Name       string
	ID         [32]byte
	Status     uint8
	BinaryURL  string
	ConfigURL  string
	SecretsURL string
}

func registerWorkflow(
	t *testing.T,
	th *EthBlockchain,
	wfRegC *workflow_registry_wrapper.WorkflowRegistry,
	input Workflow,
	donID uint32,
) {
	t.Helper()
	_, err := wfRegC.RegisterWorkflow(th.transactionOpts, input.Name, input.ID, donID,
		input.Status, input.BinaryURL, input.ConfigURL, input.SecretsURL)
	require.NoError(t, err, "failed to register workflow")
	th.Backend.Commit()
	th.Backend.Commit()
	th.Backend.Commit()
}

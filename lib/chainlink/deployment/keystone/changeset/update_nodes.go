package changeset

import (
	"errors"
	"fmt"
	"time"

	chainsel "github.com/smartcontractkit/chain-selectors"
	"github.com/smartcontractkit/mcms"
	"github.com/smartcontractkit/mcms/sdk"
	"github.com/smartcontractkit/mcms/types"

	"github.com/smartcontractkit/chainlink/deployment"
	"github.com/smartcontractkit/chainlink/deployment/common/proposalutils"

	"github.com/smartcontractkit/chainlink/deployment/keystone/changeset/internal"
	"github.com/smartcontractkit/chainlink/v2/core/services/keystore/keys/p2pkey"
)

type MCMSConfig struct {
	MinDuration time.Duration
}

var _ deployment.ChangeSet[*UpdateNodesRequest] = UpdateNodes

type UpdateNodesRequest struct {
	RegistryChainSel uint64
	P2pToUpdates     map[p2pkey.PeerID]NodeUpdate

	// MCMSConfig is optional. If non-nil, the changes will be proposed using MCMS.
	MCMSConfig *MCMSConfig
}

func (r *UpdateNodesRequest) Validate(e deployment.Environment) error {
	if r.P2pToUpdates == nil {
		return errors.New("P2pToUpdates must be non-nil")
	}

	_, exists := chainsel.ChainBySelector(r.RegistryChainSel)
	if !exists {
		return fmt.Errorf("invalid registry chain selector %d: selector does not exist", r.RegistryChainSel)
	}

	_, exists = e.Chains[r.RegistryChainSel]
	if !exists {
		return fmt.Errorf("invalid registry chain selector %d: chain does not exist in environment", r.RegistryChainSel)
	}

	return nil
}

func (r UpdateNodesRequest) UseMCMS() bool {
	return r.MCMSConfig != nil
}

type NodeUpdate = internal.NodeUpdate

// UpdateNodes updates a set of nodes.
// The nodes and capabilities in the request must already exist in the registry contract.
func UpdateNodes(env deployment.Environment, req *UpdateNodesRequest) (deployment.ChangesetOutput, error) {
	// extract the registry contract and chain from the environment
	registryChain, ok := env.Chains[req.RegistryChainSel]
	if !ok {
		return deployment.ChangesetOutput{}, fmt.Errorf("registry chain selector %d does not exist in environment", req.RegistryChainSel)
	}
	cresp, err := GetContractSets(env.Logger, &GetContractSetsRequest{
		Chains:      env.Chains,
		AddressBook: env.ExistingAddresses,
	})
	if err != nil {
		return deployment.ChangesetOutput{}, fmt.Errorf("failed to get contract sets: %w", err)
	}
	contracts, exists := cresp.ContractSets[req.RegistryChainSel]
	if !exists {
		return deployment.ChangesetOutput{}, fmt.Errorf("contract set not found for chain %d", req.RegistryChainSel)
	}

	resp, err := internal.UpdateNodes(env.Logger, &internal.UpdateNodesRequest{
		Chain:                registryChain,
		CapabilitiesRegistry: contracts.CapabilitiesRegistry,
		P2pToUpdates:         req.P2pToUpdates,
		UseMCMS:              req.UseMCMS(),
	})
	if err != nil {
		return deployment.ChangesetOutput{}, fmt.Errorf("failed to update don: %w", err)
	}

	out := deployment.ChangesetOutput{}
	if req.UseMCMS() {
		if resp.Ops == nil {
			return out, errors.New("expected MCMS operation to be non-nil")
		}
		timelocksPerChain := map[uint64]string{
			req.RegistryChainSel: contracts.Timelock.Address().Hex(),
		}
		proposerMCMSes := map[uint64]string{
			req.RegistryChainSel: contracts.ProposerMcm.Address().Hex(),
		}
		inspector, err := proposalutils.McmsInspectorForChain(env, req.RegistryChainSel)
		if err != nil {
			return deployment.ChangesetOutput{}, err
		}
		inspectorPerChain := map[uint64]sdk.Inspector{
			req.RegistryChainSel: inspector,
		}

		proposal, err := proposalutils.BuildProposalFromBatchesV2(
			env,
			timelocksPerChain,
			proposerMCMSes,
			inspectorPerChain,
			[]types.BatchOperation{*resp.Ops},
			"proposal to set update nodes",
			req.MCMSConfig.MinDuration,
		)
		if err != nil {
			return out, fmt.Errorf("failed to build proposal: %w", err)
		}
		out.MCMSTimelockProposals = []mcms.TimelockProposal{*proposal}
	}

	return out, nil
}

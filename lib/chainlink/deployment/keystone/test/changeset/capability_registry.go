package changeset

import (
	"fmt"
	"testing"

	chainsel "github.com/smartcontractkit/chain-selectors"

	"github.com/smartcontractkit/chainlink/deployment"

	"github.com/smartcontractkit/chainlink/deployment/common/view/v1_0"

	"github.com/smartcontractkit/chainlink/deployment/keystone/changeset"
	capabilities_registry "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/keystone/generated/capabilities_registry_1_1_0"
)

type HydrateConfig struct {
	ChainID uint64
}

// HydrateCapabilityRegistry deploys a new capabilities registry contract and hydrates it with the provided data.
func HydrateCapabilityRegistry(t *testing.T, v v1_0.CapabilityRegistryView, env deployment.Environment, cfg HydrateConfig) (*capabilities_registry.CapabilitiesRegistry, error) {
	t.Helper()
	chainSelector, err := chainsel.SelectorFromChainId(cfg.ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain selector from chain id: %w", err)
	}
	chain, ok := env.Chains[chainSelector]
	if !ok {
		return nil, fmt.Errorf("chain with id %d not found", cfg.ChainID)
	}
	changesetOutput, err := changeset.DeployCapabilityRegistry(env, chainSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy contract: %w", err)
	}

	resp, err := changeset.GetContractSets(env.Logger, &changeset.GetContractSetsRequest{
		Chains:      env.Chains,
		AddressBook: changesetOutput.AddressBook,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get contract sets: %w", err)
	}
	cs, ok := resp.ContractSets[chainSelector]
	if !ok {
		return nil, fmt.Errorf("failed to get contract set for chain selector: %d, chain ID: %d", chainSelector, cfg.ChainID)
	}

	deployedContract := cs.CapabilitiesRegistry

	nopsParams := v.NopsToNopsParams()
	tx, err := deployedContract.AddNodeOperators(chain.DeployerKey, nopsParams)
	if _, err = deployment.ConfirmIfNoError(chain, tx, deployment.DecodeErr(capabilities_registry.CapabilitiesRegistryABI, err)); err != nil {
		return nil, fmt.Errorf("failed to add node operators: %w", err)
	}

	capabilitiesParams := v.CapabilitiesToCapabilitiesParams()
	tx, err = deployedContract.AddCapabilities(chain.DeployerKey, capabilitiesParams)
	if _, err = deployment.ConfirmIfNoError(chain, tx, deployment.DecodeErr(capabilities_registry.CapabilitiesRegistryABI, err)); err != nil {
		return nil, fmt.Errorf("failed to add capabilities: %w", err)
	}

	nodesParams, err := v.NodesToNodesParams()
	if err != nil {
		return nil, fmt.Errorf("failed to convert nodes to nodes params: %w", err)
	}
	tx, err = deployedContract.AddNodes(chain.DeployerKey, nodesParams)
	if _, err = deployment.ConfirmIfNoError(chain, tx, deployment.DecodeErr(capabilities_registry.CapabilitiesRegistryABI, err)); err != nil {
		return nil, fmt.Errorf("failed to add nodes: %w", err)
	}

	for _, don := range v.Dons {
		cfgs, err := v.CapabilityConfigToCapabilityConfigParams(don)
		if err != nil {
			return nil, fmt.Errorf("failed to convert capability configurations to capability configuration params: %w", err)
		}
		var peerIds [][32]byte
		for _, id := range don.NodeP2PIds {
			peerIds = append(peerIds, id)
		}
		tx, err = deployedContract.AddDON(chain.DeployerKey, peerIds, cfgs, don.IsPublic, don.AcceptsWorkflows, don.F)
		if _, err = deployment.ConfirmIfNoError(chain, tx, deployment.DecodeErr(capabilities_registry.CapabilitiesRegistryABI, err)); err != nil {
			return nil, fmt.Errorf("failed to add don: %w", err)
		}
	}

	return deployedContract, nil
}

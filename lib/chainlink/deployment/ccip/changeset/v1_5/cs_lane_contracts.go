package v1_5

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/smartcontractkit/chainlink/deployment"
	"github.com/smartcontractkit/chainlink/deployment/ccip/changeset"
	price_registry_1_2_0 "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/ccip/generated/v1_2_0/price_registry"
	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/ccip/generated/v1_2_0/router"
	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/ccip/generated/v1_5_0/commit_store"
	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/ccip/generated/v1_5_0/evm_2_evm_offramp"
	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/ccip/generated/v1_5_0/evm_2_evm_onramp"
)

var _ deployment.ChangeSet[DeployLanesConfig] = DeployLanesChangeset

type DeployLanesConfig struct {
	Configs []DeployLaneConfig
}

func (c *DeployLanesConfig) Validate(e deployment.Environment, state changeset.CCIPOnChainState) error {
	for _, cfg := range c.Configs {
		if err := cfg.Validate(e, state); err != nil {
			return err
		}
	}
	return nil
}

type DeployLaneConfig struct {
	SourceChainSelector      uint64
	DestinationChainSelector uint64

	// onRamp specific configuration
	OnRampStaticCfg         evm_2_evm_onramp.EVM2EVMOnRampStaticConfig
	OnRampDynamicCfg        evm_2_evm_onramp.EVM2EVMOnRampDynamicConfig
	OnRampFeeTokenArgs      []evm_2_evm_onramp.EVM2EVMOnRampFeeTokenConfigArgs
	OnRampTransferTokenCfgs []evm_2_evm_onramp.EVM2EVMOnRampTokenTransferFeeConfigArgs
	OnRampNopsAndWeight     []evm_2_evm_onramp.EVM2EVMOnRampNopAndWeight
	OnRampRateLimiterCfg    evm_2_evm_onramp.RateLimiterConfig

	// offRamp specific configuration
	OffRampRateLimiterCfg evm_2_evm_offramp.RateLimiterConfig

	// Price Registry specific configuration
	InitialTokenPrices []price_registry_1_2_0.InternalTokenPriceUpdate
	GasPriceUpdates    []price_registry_1_2_0.InternalGasPriceUpdate
}

func (c *DeployLaneConfig) Validate(e deployment.Environment, state changeset.CCIPOnChainState) error {
	if err := deployment.IsValidChainSelector(c.SourceChainSelector); err != nil {
		return err
	}
	if err := deployment.IsValidChainSelector(c.DestinationChainSelector); err != nil {
		return err
	}
	sourceChain, exists := e.Chains[c.SourceChainSelector]
	if !exists {
		return fmt.Errorf("source chain %d not found in environment", c.SourceChainSelector)
	}
	destChain, exists := e.Chains[c.DestinationChainSelector]
	if !exists {
		return fmt.Errorf("destination chain %d not found in environment", c.DestinationChainSelector)
	}
	sourceChainState, exists := state.Chains[c.SourceChainSelector]
	if !exists {
		return fmt.Errorf("source chain %d not found in state", c.SourceChainSelector)
	}
	destChainState, exists := state.Chains[c.DestinationChainSelector]
	if !exists {
		return fmt.Errorf("destination chain %d not found in state", c.DestinationChainSelector)
	}
	// check for existing chain contracts on both source and destination chains
	if err := arePrerequisitesMet(sourceChainState, sourceChain); err != nil {
		return err
	}
	if err := arePrerequisitesMet(destChainState, destChain); err != nil {
		return err
	}
	// TODO: Add rest of the config validation
	return nil
}

func (c *DeployLaneConfig) populateAddresses(state changeset.CCIPOnChainState) error {
	sourceChainState := state.Chains[c.SourceChainSelector]
	srcLink, err := sourceChainState.LinkTokenAddress()
	if err != nil {
		return fmt.Errorf("failed to get LINK token address for source chain %d: %w", c.SourceChainSelector, err)
	}
	c.OnRampStaticCfg.LinkToken = srcLink
	c.OnRampStaticCfg.RmnProxy = sourceChainState.RMNProxy.Address()
	c.OnRampStaticCfg.TokenAdminRegistry = sourceChainState.TokenAdminRegistry.Address()

	c.OnRampDynamicCfg.Router = sourceChainState.Router.Address()
	c.OnRampDynamicCfg.PriceRegistry = sourceChainState.PriceRegistry.Address()
	return nil
}

func DeployLanesChangeset(env deployment.Environment, c DeployLanesConfig) (deployment.ChangesetOutput, error) {
	state, err := changeset.LoadOnchainState(env)
	if err != nil {
		return deployment.ChangesetOutput{}, fmt.Errorf("failed to load CCIP onchain state: %w", err)
	}
	if err := c.Validate(env, state); err != nil {
		return deployment.ChangesetOutput{}, fmt.Errorf("invalid DeployChainContractsConfig: %w", err)
	}
	// populate addresses from the state
	for i := range c.Configs {
		if err := c.Configs[i].populateAddresses(state); err != nil {
			return deployment.ChangesetOutput{}, err
		}
	}
	newAddresses := deployment.NewMemoryAddressBook()
	for _, cfg := range c.Configs {
		if err := deployLane(env, state, newAddresses, cfg); err != nil {
			return deployment.ChangesetOutput{
				AddressBook: newAddresses,
			}, err
		}
	}
	return deployment.ChangesetOutput{
		AddressBook: newAddresses,
	}, nil
}

func deployLane(e deployment.Environment, state changeset.CCIPOnChainState, ab deployment.AddressBook, cfg DeployLaneConfig) error {
	// update prices on the source price registry
	sourceChainState := state.Chains[cfg.SourceChainSelector]
	destChainState := state.Chains[cfg.DestinationChainSelector]
	sourceChain := e.Chains[cfg.SourceChainSelector]
	destChain := e.Chains[cfg.DestinationChainSelector]
	sourcePriceReg := sourceChainState.PriceRegistry
	tx, err := sourcePriceReg.UpdatePrices(sourceChain.DeployerKey, price_registry_1_2_0.InternalPriceUpdates{
		TokenPriceUpdates: cfg.InitialTokenPrices,
		GasPriceUpdates:   cfg.GasPriceUpdates,
	})
	if err != nil {
		return err
	}
	_, err = sourceChain.Confirm(tx)
	if err != nil {
		return fmt.Errorf("failed to confirm price update tx for chain %s: %w", sourceChain.String(), deployment.MaybeDataErr(err))
	}
	// ================================================================
	// │                        Deploy Lane                           │
	// ================================================================
	// Deploy onRamp on source chain
	onRamp, onRampExists := sourceChainState.EVM2EVMOnRamp[cfg.DestinationChainSelector]
	if !onRampExists {
		onRampC, err := deployment.DeployContract(e.Logger, sourceChain, ab,
			func(chain deployment.Chain) deployment.ContractDeploy[*evm_2_evm_onramp.EVM2EVMOnRamp] {
				onRampAddress, tx2, onRampC, err2 := evm_2_evm_onramp.DeployEVM2EVMOnRamp(
					sourceChain.DeployerKey,
					sourceChain.Client,
					cfg.OnRampStaticCfg,
					cfg.OnRampDynamicCfg,
					cfg.OnRampRateLimiterCfg,
					cfg.OnRampFeeTokenArgs,
					cfg.OnRampTransferTokenCfgs,
					cfg.OnRampNopsAndWeight,
				)
				return deployment.ContractDeploy[*evm_2_evm_onramp.EVM2EVMOnRamp]{
					Address: onRampAddress, Contract: onRampC, Tx: tx2,
					Tv: deployment.NewTypeAndVersion(changeset.OnRamp, deployment.Version1_5_0), Err: err2,
				}
			})
		if err != nil {
			e.Logger.Errorw("Failed to deploy EVM2EVMOnRamp", "chain", sourceChain.String(), "err", err)
			return err
		}
		onRamp = onRampC.Contract
	} else {
		e.Logger.Infow("EVM2EVMOnRamp already exists",
			"source chain", sourceChain.String(), "destination chain", destChain.String(),
			"address", onRamp.Address().String())
	}

	// Deploy commit store on source chain
	commitStore, commitStoreExists := destChainState.CommitStore[cfg.SourceChainSelector]
	if !commitStoreExists {
		commitStoreC, err := deployment.DeployContract(e.Logger, destChain, ab,
			func(chain deployment.Chain) deployment.ContractDeploy[*commit_store.CommitStore] {
				commitStoreAddress, tx2, commitStoreC, err2 := commit_store.DeployCommitStore(
					destChain.DeployerKey,
					destChain.Client,
					commit_store.CommitStoreStaticConfig{
						ChainSelector:       destChain.Selector,
						SourceChainSelector: sourceChain.Selector,
						OnRamp:              onRamp.Address(),
						RmnProxy:            destChainState.RMNProxy.Address(),
					},
				)
				return deployment.ContractDeploy[*commit_store.CommitStore]{
					Address: commitStoreAddress, Contract: commitStoreC, Tx: tx2,
					Tv: deployment.NewTypeAndVersion(changeset.CommitStore, deployment.Version1_5_0), Err: err2,
				}
			})
		if err != nil {
			e.Logger.Errorw("Failed to deploy CommitStore", "chain", sourceChain.String(), "err", err)
			return err
		}
		commitStore = commitStoreC.Contract
	} else {
		e.Logger.Infow("CommitStore already exists",
			"source chain", sourceChain.String(), "destination chain", destChain.String(),
			"address", commitStore.Address().String())
	}

	// Deploy offRamp on destination chain
	offRamp, offRampExists := destChainState.EVM2EVMOffRamp[cfg.SourceChainSelector]
	if !offRampExists {
		offRampC, err := deployment.DeployContract(e.Logger, destChain, ab,
			func(chain deployment.Chain) deployment.ContractDeploy[*evm_2_evm_offramp.EVM2EVMOffRamp] {
				offRampAddress, tx2, offRampC, err2 := evm_2_evm_offramp.DeployEVM2EVMOffRamp(
					destChain.DeployerKey,
					destChain.Client,
					evm_2_evm_offramp.EVM2EVMOffRampStaticConfig{
						CommitStore:         commitStore.Address(),
						ChainSelector:       destChain.Selector,
						SourceChainSelector: sourceChain.Selector,
						OnRamp:              onRamp.Address(),
						PrevOffRamp:         common.HexToAddress(""),
						RmnProxy:            destChainState.RMNProxy.Address(), // RMN, formerly ARM
						TokenAdminRegistry:  destChainState.TokenAdminRegistry.Address(),
					},
					cfg.OffRampRateLimiterCfg,
				)
				return deployment.ContractDeploy[*evm_2_evm_offramp.EVM2EVMOffRamp]{
					Address: offRampAddress, Contract: offRampC, Tx: tx2,
					Tv: deployment.NewTypeAndVersion(changeset.OffRamp, deployment.Version1_5_0), Err: err2,
				}
			})
		if err != nil {
			e.Logger.Errorw("Failed to deploy EVM2EVMOffRamp", "chain", sourceChain.String(), "err", err)
			return err
		}
		offRamp = offRampC.Contract
	} else {
		e.Logger.Infow("EVM2EVMOffRamp already exists",
			"source chain", sourceChain.String(), "destination chain", destChain.String(),
			"address", offRamp.Address().String())
	}

	// Apply Router updates
	tx, err = sourceChainState.Router.ApplyRampUpdates(sourceChain.DeployerKey,
		[]router.RouterOnRamp{{DestChainSelector: destChain.Selector, OnRamp: onRamp.Address()}}, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to apply router updates for source chain %s: %w", sourceChain.String(), deployment.MaybeDataErr(err))
	}
	_, err = sourceChain.Confirm(tx)
	if err != nil {
		return fmt.Errorf("failed to confirm router updates tx %s for source chain %s: %w", tx.Hash().String(), sourceChain.String(), deployment.MaybeDataErr(err))
	}

	tx, err = destChainState.Router.ApplyRampUpdates(destChain.DeployerKey,
		nil,
		nil,
		[]router.RouterOffRamp{{SourceChainSelector: sourceChain.Selector, OffRamp: offRamp.Address()}},
	)
	if err != nil {
		return fmt.Errorf("failed to apply router updates for destination chain %s: %w", destChain.String(), deployment.MaybeDataErr(err))
	}
	_, err = destChain.Confirm(tx)
	if err != nil {
		return fmt.Errorf("failed to confirm router updates tx %s for destination chain %s: %w", tx.Hash().String(), destChain.String(), deployment.MaybeDataErr(err))
	}

	// price registry updates
	_, err = destChainState.PriceRegistry.ApplyPriceUpdatersUpdates(
		destChain.DeployerKey,
		[]common.Address{commitStore.Address()},
		[]common.Address{},
	)
	if err != nil {
		return fmt.Errorf("failed to apply price registry updates for destination chain %s: %w", destChain.String(), deployment.MaybeDataErr(err))
	}
	_, err = destChain.Confirm(tx)
	if err != nil {
		return fmt.Errorf("failed to confirm price registry updates tx %s for destination chain %s: %w", tx.Hash().String(), destChain.String(), deployment.MaybeDataErr(err))
	}
	return nil
}

func arePrerequisitesMet(chainState changeset.CCIPChainState, chain deployment.Chain) error {
	if chainState.Router == nil {
		return fmt.Errorf("router not found for chain %s", chain.String())
	}
	if chainState.PriceRegistry == nil {
		return fmt.Errorf("price registry not found for chain %s", chain.String())
	}
	if chainState.RMN == nil && chainState.MockRMN == nil {
		return fmt.Errorf("neither RMN nor mockRMN found for chain %s", chain.String())
	}
	if chainState.Weth9 == nil {
		return fmt.Errorf("WETH9 not found for chain %s", chain.String())
	}
	if _, err := chainState.LinkTokenAddress(); err != nil {
		return fmt.Errorf("LINK token not found for chain %s", chain.String())
	}
	if chainState.TokenAdminRegistry == nil {
		return fmt.Errorf("token admin registry not found for chain %s", chain.String())
	}
	if chainState.RMNProxy == nil {
		return fmt.Errorf("RMNProxy not found for chain %s", chain.String())
	}
	return nil
}

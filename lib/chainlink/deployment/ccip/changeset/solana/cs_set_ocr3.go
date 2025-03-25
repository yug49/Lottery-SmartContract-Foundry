package solana

import (
	"errors"
	"fmt"

	"github.com/gagliardetto/solana-go"
	chain_selectors "github.com/smartcontractkit/chain-selectors"
	"github.com/smartcontractkit/mcms"
	"github.com/smartcontractkit/mcms/sdk"
	mcmsSolana "github.com/smartcontractkit/mcms/sdk/solana"
	mcmsTypes "github.com/smartcontractkit/mcms/types"

	solOffRamp "github.com/smartcontractkit/chainlink-ccip/chains/solana/gobindings/ccip_offramp"

	"github.com/smartcontractkit/chainlink/deployment"
	"github.com/smartcontractkit/chainlink/deployment/ccip/changeset"
	ccipChangeset "github.com/smartcontractkit/chainlink/deployment/ccip/changeset"
	"github.com/smartcontractkit/chainlink/deployment/ccip/changeset/internal"
	"github.com/smartcontractkit/chainlink/deployment/ccip/changeset/v1_6"
	csState "github.com/smartcontractkit/chainlink/deployment/common/changeset/state"
	"github.com/smartcontractkit/chainlink/deployment/common/proposalutils"
)

const (
	OcrCommitPlugin uint8 = iota
	OcrExecutePlugin
)

// SET OCR3 CONFIG
func btoi(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

// SetOCR3OffRamp will set the OCR3 offramp for the given chain.
// to the active configuration on CCIPHome. This
// is used to complete the candidate->active promotion cycle, it's
// run after the candidate is confirmed to be working correctly.
// Multichain is especially helpful for NOP rotations where we have
// to touch all the chain to change signers.
func SetOCR3ConfigSolana(e deployment.Environment, cfg v1_6.SetOCR3OffRampConfig) (deployment.ChangesetOutput, error) {
	state, err := changeset.LoadOnchainState(e)
	if err != nil {
		return deployment.ChangesetOutput{}, fmt.Errorf("failed to load onchain state: %w", err)
	}

	if err := cfg.Validate(e, state); err != nil {
		return deployment.ChangesetOutput{}, err
	}

	for _, remote := range cfg.RemoteChainSels {
		chainFamily, _ := chain_selectors.GetSelectorFamily(remote)
		if chainFamily != chain_selectors.FamilySolana {
			return deployment.ChangesetOutput{}, fmt.Errorf("chain %d is not a solana chain", remote)
		}
		chain := e.SolChains[remote]
		if err := ccipChangeset.ValidateOwnershipSolana(&e, chain, cfg.MCMS != nil, state.SolChains[remote].OffRamp, ccipChangeset.OffRamp, solana.PublicKey{}); err != nil {
			return deployment.ChangesetOutput{}, fmt.Errorf("failed to validate ownership: %w", err)
		}
	}

	timelocks := map[uint64]string{}
	proposers := map[uint64]string{}
	inspectors := map[uint64]sdk.Inspector{}
	var batches []mcmsTypes.BatchOperation
	for _, remote := range cfg.RemoteChainSels {
		donID, err := internal.DonIDForChain(
			state.Chains[cfg.HomeChainSel].CapabilityRegistry,
			state.Chains[cfg.HomeChainSel].CCIPHome,
			remote)
		if err != nil {
			return deployment.ChangesetOutput{}, fmt.Errorf("failed to get don id for chain %d: %w", remote, err)
		}
		args, err := internal.BuildSetOCR3ConfigArgsSolana(donID, state.Chains[cfg.HomeChainSel].CCIPHome, remote)
		if err != nil {
			return deployment.ChangesetOutput{}, fmt.Errorf("failed to build set ocr3 config args: %w", err)
		}
		set, err := isOCR3ConfigSetOnOffRampSolana(e, e.SolChains[remote], state.SolChains[remote], args)
		if err != nil {
			return deployment.ChangesetOutput{}, fmt.Errorf("failed to check if ocr3 config is set on offramp: %w", err)
		}
		if set {
			e.Logger.Infof("OCR3 config already set on offramp for chain %d", remote)
			continue
		}
		chain := e.SolChains[remote]
		addresses, _ := e.ExistingAddresses.AddressesForChain(remote)
		mcmState, _ := csState.MaybeLoadMCMSWithTimelockChainStateSolana(chain, addresses)

		timelocks[remote] = mcmsSolana.ContractAddress(
			mcmState.TimelockProgram,
			mcmsSolana.PDASeed(mcmState.TimelockSeed),
		)
		proposers[remote] = mcmsSolana.ContractAddress(mcmState.McmProgram, mcmsSolana.PDASeed(mcmState.ProposerMcmSeed))
		inspectors[remote] = mcmsSolana.NewInspector(chain.Client)

		var instructions []solana.Instruction
		var txns []mcmsTypes.Transaction
		offRampConfigPDA := state.SolChains[remote].OffRampConfigPDA
		offRampStatePDA := state.SolChains[remote].OffRampStatePDA
		solOffRamp.SetProgramID(state.SolChains[remote].OffRamp)
		var authority solana.PublicKey
		if cfg.MCMS != nil {
			authority, err = FetchTimelockSigner(e, remote)
			if err != nil {
				return deployment.ChangesetOutput{}, fmt.Errorf("failed to fetch timelock signer: %w", err)
			}
		} else {
			authority = e.SolChains[remote].DeployerKey.PublicKey()
		}
		for _, arg := range args {
			var ocrType solOffRamp.OcrPluginType
			switch arg.OCRPluginType {
			case OcrCommitPlugin:
				ocrType = solOffRamp.Commit_OcrPluginType
			case OcrExecutePlugin:
				ocrType = solOffRamp.Execution_OcrPluginType
			default:
				return deployment.ChangesetOutput{}, errors.New("invalid OCR plugin type")
			}
			instruction, err := solOffRamp.NewSetOcrConfigInstruction(
				ocrType,
				solOffRamp.Ocr3ConfigInfo{
					ConfigDigest:                   arg.ConfigDigest,
					F:                              arg.F,
					IsSignatureVerificationEnabled: btoi(arg.IsSignatureVerificationEnabled),
				},
				arg.Signers,
				arg.Transmitters,
				offRampConfigPDA,
				offRampStatePDA,
				authority,
			).ValidateAndBuild()
			if err != nil {
				return deployment.ChangesetOutput{}, fmt.Errorf("failed to generate instructions: %w", err)
			}
			if cfg.MCMS == nil {
				instructions = append(instructions, instruction)
			} else {
				tx, err := BuildMCMSTxn(instruction, state.SolChains[remote].OffRamp.String(), ccipChangeset.OffRamp)
				if err != nil {
					return deployment.ChangesetOutput{}, fmt.Errorf("failed to create transaction: %w", err)
				}
				txns = append(txns, *tx)
			}
		}
		if cfg.MCMS == nil {
			if err := e.SolChains[remote].Confirm(instructions); err != nil {
				return deployment.ChangesetOutput{}, fmt.Errorf("failed to confirm instructions: %w", err)
			}
		} else {
			batches = append(batches, mcmsTypes.BatchOperation{
				ChainSelector: mcmsTypes.ChainSelector(remote),
				Transactions:  txns,
			})
		}
	}
	if cfg.MCMS != nil {
		proposal, err := proposalutils.BuildProposalFromBatchesV2(
			e,
			timelocks,
			proposers,
			inspectors,
			batches,
			"set ocr3 config for Solana",
			cfg.MCMS.MinDelay,
		)
		if err != nil {
			return deployment.ChangesetOutput{}, fmt.Errorf("failed to build proposal: %w", err)
		}
		return deployment.ChangesetOutput{
			MCMSTimelockProposals: []mcms.TimelockProposal{*proposal},
		}, nil
	}
	return deployment.ChangesetOutput{}, nil
}

func isOCR3ConfigSetOnOffRampSolana(
	e deployment.Environment,
	chain deployment.SolChain,
	chainState ccipChangeset.SolCCIPChainState,
	args []internal.MultiOCR3BaseOCRConfigArgsSolana,
) (bool, error) {
	var configAccount solOffRamp.Config
	err := chain.GetAccountDataBorshInto(e.GetContext(), chainState.OffRampConfigPDA, &configAccount)
	if err != nil {
		return false, fmt.Errorf("failed to get account info: %w", err)
	}
	for _, newState := range args {
		existingState := configAccount.Ocr3[newState.OCRPluginType]
		if existingState.ConfigInfo.ConfigDigest != newState.ConfigDigest {
			e.Logger.Infof("OCR3 config digest mismatch")
			return false, nil
		}
		if existingState.ConfigInfo.F != newState.F {
			e.Logger.Infof("OCR3 config F mismatch")
			return false, nil
		}
		if existingState.ConfigInfo.IsSignatureVerificationEnabled != btoi(newState.IsSignatureVerificationEnabled) {
			e.Logger.Infof("OCR3 config signature verification mismatch")
			return false, nil
		}
		if newState.OCRPluginType == OcrCommitPlugin {
			// only commit will set signers, exec doesn't need them.
			if len(existingState.Signers) != len(newState.Signers) {
				e.Logger.Infof("OCR3 config signers length mismatch")
				return false, nil
			}
			for i := 0; i < len(existingState.Signers); i++ {
				if existingState.Signers[i] != newState.Signers[i] {
					e.Logger.Infof("OCR3 config signers mismatch")
					return false, nil
				}
			}
		}
		if len(existingState.Transmitters) != len(newState.Transmitters) {
			e.Logger.Infof("OCR3 config transmitters length mismatch")
			return false, nil
		}
		for i := 0; i < len(existingState.Transmitters); i++ {
			if existingState.Transmitters[i] != newState.Transmitters[i] {
				e.Logger.Infof("OCR3 config transmitters mismatch")
				return false, nil
			}
		}
	}
	return true, nil
}

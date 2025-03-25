package ccipevm

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	chainsel "github.com/smartcontractkit/chain-selectors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-ccip/pkg/types/ccipocr3"
)

func Test_calculateMessageMaxGas(t *testing.T) {
	type args struct {
		dataLen          int
		numTokens        int
		extraArgs        []byte
		tokenGasOverhead uint32
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			name: "base",
			args: args{dataLen: 5, numTokens: 2, extraArgs: makeExtraArgsV1(200_000), tokenGasOverhead: 10},
			want: 922_284,
		},
		{
			name: "large",
			args: args{dataLen: 1000, numTokens: 1000, extraArgs: makeExtraArgsV1(200_000), tokenGasOverhead: 1},
			want: 296_678_520,
		},
		{
			name: "overheadGas test 1",
			args: args{dataLen: 0, numTokens: 0, extraArgs: makeExtraArgsV1(200_000), tokenGasOverhead: 100},
			want: 319_920,
		},
		{
			name: "overheadGas test 2",
			args: args{
				dataLen:          len([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0}),
				numTokens:        1,
				extraArgs:        makeExtraArgsV1(200_000),
				tokenGasOverhead: 2,
			},
			want: 625_950,
		},
		{
			name: "allowOOO set to true makes no difference to final gas estimate",
			args: args{
				dataLen:          5,
				numTokens:        2,
				extraArgs:        makeExtraArgsV2(200_000, true),
				tokenGasOverhead: 100,
			},
			want: 922_464,
		},
		{
			name: "allowOOO set to false makes no difference to final gas estimate",
			args: args{
				dataLen:          5,
				numTokens:        2,
				extraArgs:        makeExtraArgsV2(200_000, false),
				tokenGasOverhead: 100,
			},
			want: 922_464,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := ccipocr3.Message{
				Data:         make([]byte, tt.args.dataLen),
				TokenAmounts: getTokenAmounts(t, tt.args.numTokens, tt.args.tokenGasOverhead),
				ExtraArgs:    tt.args.extraArgs,
			}
			// Set the source chain selector to be EVM for now
			msg.Header.SourceChainSelector = ccipocr3.ChainSelector(chainsel.ETHEREUM_TESTNET_SEPOLIA.Selector)
			ep := EstimateProvider{extraDataCodec: ExtraDataCodec}
			got := ep.CalculateMessageMaxGas(msg)
			t.Log(got)
			assert.Equalf(t, tt.want, got, "calculateMessageMaxGas(%v, %v)", tt.args.dataLen, tt.args.numTokens)
		})
	}
}

// TestCalculateMaxGas is taken from the ccip repo where the CalculateMerkleTreeGas and CalculateMessageMaxGas values
// are combined to one function.
func TestCalculateMaxGas(t *testing.T) {
	tests := []struct {
		name             string
		numRequests      int
		dataLength       int
		numberOfTokens   int
		extraArgs        []byte
		tokenGasOverhead uint32
		want             uint64
	}{
		{
			name:             "maxGasOverheadGas 1",
			numRequests:      6,
			dataLength:       0,
			numberOfTokens:   0,
			extraArgs:        makeExtraArgsV1(200_000),
			tokenGasOverhead: 10,
			want:             322_992,
		},
		{
			name:             "maxGasOverheadGas 2",
			numRequests:      3,
			dataLength:       len([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0}),
			numberOfTokens:   1,
			extraArgs:        makeExtraArgsV1(200_000),
			tokenGasOverhead: 10,
			want:             628_518,
		},
		{
			name:             "v2 extra args",
			numRequests:      3,
			dataLength:       len([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0}),
			numberOfTokens:   1,
			extraArgs:        makeExtraArgsV2(200_000, true),
			tokenGasOverhead: 10,
			want:             628_518,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := ccipocr3.Message{
				Data:         make([]byte, tt.dataLength),
				TokenAmounts: getTokenAmounts(t, tt.numberOfTokens, tt.tokenGasOverhead),
				ExtraArgs:    tt.extraArgs,
			}

			msg.Header.SourceChainSelector = ccipocr3.ChainSelector(chainsel.ETHEREUM_TESTNET_SEPOLIA.Selector)
			ep := EstimateProvider{extraDataCodec: ExtraDataCodec}
			gotTree := ep.CalculateMerkleTreeGas(tt.numRequests)
			gotMsg := ep.CalculateMessageMaxGas(msg)
			t.Log("want", tt.want, "got", gotTree+gotMsg)
			assert.Equal(t, tt.want, gotTree+gotMsg)
		})
	}
}

func makeExtraArgsV1(gasLimit uint64) []byte {
	// extra args is the tag followed by the gas limit abi-encoded.
	var extraArgs []byte
	extraArgs = append(extraArgs, evmExtraArgsV1Tag...)
	gasLimitBytes := new(big.Int).SetUint64(gasLimit).Bytes()
	// pad from the left to 32 bytes
	gasLimitBytes = common.LeftPadBytes(gasLimitBytes, 32)
	extraArgs = append(extraArgs, gasLimitBytes...)
	return extraArgs
}

func makeExtraArgsV2(gasLimit uint64, allowOOO bool) []byte {
	// extra args is the tag followed by the gas limit and allowOOO abi-encoded.
	var extraArgs []byte
	extraArgs = append(extraArgs, evmExtraArgsV2Tag...)
	gasLimitBytes := new(big.Int).SetUint64(gasLimit).Bytes()
	// pad from the left to 32 bytes
	gasLimitBytes = common.LeftPadBytes(gasLimitBytes, 32)

	// abi-encode allowOOO
	var allowOOOBytes []byte
	if allowOOO {
		allowOOOBytes = append(allowOOOBytes, 1)
	} else {
		allowOOOBytes = append(allowOOOBytes, 0)
	}
	// pad from the left to 32 bytes
	allowOOOBytes = common.LeftPadBytes(allowOOOBytes, 32)

	extraArgs = append(extraArgs, gasLimitBytes...)
	extraArgs = append(extraArgs, allowOOOBytes...)
	return extraArgs
}

func getTokenAmounts(t *testing.T, numTokens int, tokenGasOverhead uint32) []ccipocr3.RampTokenAmount {
	tokenDestGasOverhead, err := abiEncodeUint32(tokenGasOverhead)
	require.NoError(t, err)

	tokenAmounts := make([]ccipocr3.RampTokenAmount, numTokens)
	for i := 0; i < numTokens; i++ {
		tokenAmounts[i] = ccipocr3.RampTokenAmount{
			DestExecData: tokenDestGasOverhead,
		}
	}
	return tokenAmounts
}

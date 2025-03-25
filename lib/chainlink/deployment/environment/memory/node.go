package memory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/maps"

	chainsel "github.com/smartcontractkit/chain-selectors"

	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/smartcontractkit/chainlink-integrations/evm/keys"
	nodev1 "github.com/smartcontractkit/chainlink-protos/job-distributor/v1/node"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	mnCfg "github.com/smartcontractkit/chainlink-framework/multinode/config"

	solcfg "github.com/smartcontractkit/chainlink-solana/pkg/solana/config"

	"github.com/smartcontractkit/chainlink/deployment"
	"github.com/smartcontractkit/chainlink/deployment/logger"

	"github.com/smartcontractkit/chainlink-integrations/evm/assets"
	"github.com/smartcontractkit/chainlink-integrations/evm/client"
	v2toml "github.com/smartcontractkit/chainlink-integrations/evm/config/toml"
	"github.com/smartcontractkit/chainlink-integrations/evm/testutils"
	evmutils "github.com/smartcontractkit/chainlink-integrations/evm/utils/big"
	pb "github.com/smartcontractkit/chainlink-protos/orchestrator/feedsmanager"

	"github.com/smartcontractkit/chainlink/v2/core/capabilities"
	configv2 "github.com/smartcontractkit/chainlink/v2/core/config/toml"
	"github.com/smartcontractkit/chainlink/v2/core/logger/audit"
	"github.com/smartcontractkit/chainlink/v2/core/services/chainlink"
	feeds2 "github.com/smartcontractkit/chainlink/v2/core/services/feeds"
	feedsMocks "github.com/smartcontractkit/chainlink/v2/core/services/feeds/mocks"
	"github.com/smartcontractkit/chainlink/v2/core/services/keystore"
	"github.com/smartcontractkit/chainlink/v2/core/services/keystore/chaintype"
	"github.com/smartcontractkit/chainlink/v2/core/services/keystore/keys/csakey"
	"github.com/smartcontractkit/chainlink/v2/core/services/keystore/keys/ocr2key"
	"github.com/smartcontractkit/chainlink/v2/core/services/keystore/keys/p2pkey"
	"github.com/smartcontractkit/chainlink/v2/core/services/keystore/keys/workflowkey"
	"github.com/smartcontractkit/chainlink/v2/core/services/relay"
	"github.com/smartcontractkit/chainlink/v2/core/utils"
	"github.com/smartcontractkit/chainlink/v2/core/utils/crypto"
	"github.com/smartcontractkit/chainlink/v2/core/utils/testutils/heavyweight"
)

type Node struct {
	App chainlink.Application
	// Transmitter key/OCR keys for this node
	Chains     []uint64 // chain selectors
	Keys       Keys
	Addr       net.TCPAddr
	IsBoostrap bool
}

func (n Node) MultiAddr() string {
	a := ""
	if n.IsBoostrap {
		a = fmt.Sprintf("%s@%s", strings.TrimPrefix(n.Keys.PeerID.String(), "p2p_"), n.Addr.String())
	}
	return a
}

func (n Node) ReplayLogs(chains map[uint64]uint64) error {
	for sel, block := range chains {
		chainID, _ := chainsel.ChainIdFromSelector(sel)
		if err := n.App.ReplayFromBlock(big.NewInt(int64(chainID)), block, false); err != nil {
			return err
		}
	}
	return nil
}

// DeploymentNode is an adapter for deployment.Node
func (n Node) DeploymentNode() (deployment.Node, error) {
	jdChainConfigs, err := n.JDChainConfigs()
	if err != nil {
		return deployment.Node{}, err
	}
	selMap, err := deployment.ChainConfigsToOCRConfig(jdChainConfigs)
	if err != nil {
		return deployment.Node{}, err
	}
	// arbitrarily set the first evm chain as the transmitter
	var admin string
	for _, k := range n.Keys.Transmitters {
		admin = k
		break
	}
	return deployment.Node{
		NodeID:         n.Keys.PeerID.String(),
		Name:           n.Keys.PeerID.String(),
		SelToOCRConfig: selMap,
		CSAKey:         n.Keys.CSA.ID(),
		PeerID:         n.Keys.PeerID,
		AdminAddr:      admin,
		MultiAddr:      n.MultiAddr(),
		IsBootstrap:    n.IsBoostrap,
	}, nil
}

func (n Node) JDChainConfigs() ([]*nodev1.ChainConfig, error) {
	var chainConfigs []*nodev1.ChainConfig
	for _, selector := range n.Chains {
		family, err := chainsel.GetSelectorFamily(selector)
		if err != nil {
			return nil, err
		}

		// NOTE: this supports non-EVM too
		chainID, err := chainsel.GetChainIDFromSelector(selector)
		if err != nil {
			return nil, err
		}

		var ocrtype chaintype.ChainType
		switch family {
		case chainsel.FamilyEVM:
			ocrtype = chaintype.EVM
		case chainsel.FamilySolana:
			ocrtype = chaintype.Solana
		case chainsel.FamilyStarknet:
			ocrtype = chaintype.StarkNet
		case chainsel.FamilyCosmos:
			ocrtype = chaintype.Cosmos
		case chainsel.FamilyAptos:
			ocrtype = chaintype.Aptos
		default:
			return nil, fmt.Errorf("Unsupported chain family %v", family)
		}

		bundle := n.Keys.OCRKeyBundles[ocrtype]
		offpk := bundle.OffchainPublicKey()
		cpk := bundle.ConfigEncryptionPublicKey()

		keyBundle := &nodev1.OCR2Config_OCRKeyBundle{
			BundleId:              bundle.ID(),
			ConfigPublicKey:       common.Bytes2Hex(cpk[:]),
			OffchainPublicKey:     common.Bytes2Hex(offpk[:]),
			OnchainSigningAddress: bundle.OnChainPublicKey(),
		}

		var ctype nodev1.ChainType
		switch family {
		case chainsel.FamilyEVM:
			ctype = nodev1.ChainType_CHAIN_TYPE_EVM
		case chainsel.FamilySolana:
			ctype = nodev1.ChainType_CHAIN_TYPE_SOLANA
		case chainsel.FamilyStarknet:
			ctype = nodev1.ChainType_CHAIN_TYPE_STARKNET
		case chainsel.FamilyAptos:
			ctype = nodev1.ChainType_CHAIN_TYPE_APTOS
		default:
			panic(fmt.Sprintf("Unsupported chain family %v", family))
		}

		transmitter := n.Keys.Transmitters[selector]

		chainConfigs = append(chainConfigs, &nodev1.ChainConfig{
			Chain: &nodev1.Chain{
				Id:   chainID,
				Type: ctype,
			},
			AccountAddress: transmitter,
			AdminAddress:   transmitter,
			Ocr1Config:     nil,
			Ocr2Config: &nodev1.OCR2Config{
				Enabled:     true,
				IsBootstrap: n.IsBoostrap,
				P2PKeyBundle: &nodev1.OCR2Config_P2PKeyBundle{
					PeerId: n.Keys.PeerID.String(),
				},
				OcrKeyBundle:     keyBundle,
				Multiaddr:        n.MultiAddr(),
				Plugins:          nil, // TODO: programmatic way to list these from the embedded chainlink.Application?
				ForwarderAddress: ptr(""),
			},
		})
	}
	return chainConfigs, nil
}

type ConfigOpt func(c *chainlink.Config)

// WithFinalityDepths sets the finality depths of the evm chain
// in the map.
func WithFinalityDepths(finalityDepths map[uint64]uint32) ConfigOpt {
	return func(c *chainlink.Config) {
		for chainID, depth := range finalityDepths {
			chainIDBig := evmutils.New(new(big.Int).SetUint64(chainID))
			for _, evmChainConfig := range c.EVM {
				if evmChainConfig.ChainID.Cmp(chainIDBig) == 0 {
					evmChainConfig.Chain.FinalityDepth = ptr(depth)
				}
			}
		}
	}
}

// Creates a CL node which is:
// - Configured for OCR
// - Configured for the chains specified
// - Transmitter keys funded.
func NewNode(
	t *testing.T,
	port int, // Port for the P2P V2 listener.
	chains map[uint64]deployment.Chain,
	solchains map[uint64]deployment.SolChain,
	logLevel zapcore.Level,
	bootstrap bool,
	registryConfig deployment.CapabilityRegistryConfig,
	configOpts ...ConfigOpt,
) *Node {
	evmchains := make(map[uint64]EVMChain)
	for _, chain := range chains {
		family, err := chainsel.GetSelectorFamily(chain.Selector)
		if err != nil {
			t.Fatal(err)
		}
		// we're only mapping evm chains here, currently this list could also contain non-EVMs, e.g. Aptos
		if family != chainsel.FamilyEVM {
			continue
		}
		evmChainID, err := chainsel.ChainIdFromSelector(chain.Selector)
		if err != nil {
			t.Fatal(err)
		}
		evmchains[evmChainID] = EVMChain{
			Backend:     chain.Client.(*Backend).Sim,
			DeployerKey: chain.DeployerKey,
		}
	}

	// Do not want to load fixtures as they contain a dummy chainID.
	// Create database and initial configuration.
	cfg, db := heavyweight.FullTestDBNoFixturesV2(t, func(c *chainlink.Config, s *chainlink.Secrets) {
		c.Insecure.OCRDevelopmentMode = ptr(true) // Disables ocr spec validation so we can have fast polling for the test.

		c.Feature.LogPoller = ptr(true)

		// P2P V2 configs.
		c.P2P.V2.Enabled = ptr(true)
		c.P2P.V2.DeltaDial = config.MustNewDuration(500 * time.Millisecond)
		c.P2P.V2.DeltaReconcile = config.MustNewDuration(5 * time.Second)
		c.P2P.V2.ListenAddresses = &[]string{fmt.Sprintf("127.0.0.1:%d", port)}

		// Enable Capabilities, This is a pre-requisite for registrySyncer to work.
		if registryConfig.Contract != common.HexToAddress("0x0") {
			c.Capabilities.ExternalRegistry.NetworkID = ptr(relay.NetworkEVM)
			c.Capabilities.ExternalRegistry.ChainID = ptr(strconv.FormatUint(uint64(registryConfig.EVMChainID), 10))
			c.Capabilities.ExternalRegistry.Address = ptr(registryConfig.Contract.String())
		}

		// OCR configs
		c.OCR.Enabled = ptr(false)
		c.OCR.DefaultTransactionQueueDepth = ptr(uint32(200))
		c.OCR2.Enabled = ptr(true)
		c.OCR2.ContractPollInterval = config.MustNewDuration(5 * time.Second)

		c.Log.Level = ptr(configv2.LogLevel(logLevel))

		var evmConfigs v2toml.EVMConfigs
		for chainID := range evmchains {
			evmConfigs = append(evmConfigs, createConfigV2Chain(chainID))
		}
		c.EVM = evmConfigs

		var solConfigs solcfg.TOMLConfigs
		for chainID, chain := range solchains {
			solanaChainID, err := chainsel.GetChainIDFromSelector(chainID)
			if err != nil {
				t.Fatal(err)
			}
			solConfigs = append(solConfigs, createSolanaChainConfig(solanaChainID, chain))
		}
		c.Solana = solConfigs

		for _, opt := range configOpts {
			opt(c)
		}
	})

	// Set logging.
	lggr := logger.NewSingleFileLogger(t)
	lggr.SetLogLevel(logLevel)

	// Create clients for the core node backed by sim.
	clients := make(map[uint64]client.Client)
	for chainID, chain := range evmchains {
		clients[chainID] = client.NewSimulatedBackendClient(t, chain.Backend, big.NewInt(int64(chainID)))
	}

	master := keystore.New(db, utils.FastScryptParams, lggr)
	ctx := tests.Context(t)
	require.NoError(t, master.Unlock(ctx, "password"))
	require.NoError(t, master.CSA().EnsureKey(ctx))
	require.NoError(t, master.Workflow().EnsureKey(ctx))

	app, err := chainlink.NewApplication(ctx, chainlink.ApplicationOpts{
		CREOpts: chainlink.CREOpts{
			CapabilitiesRegistry: capabilities.NewRegistry(lggr),
		},
		Config:   cfg,
		DS:       db,
		KeyStore: master,
		EVMFactoryConfigFn: func(fc *chainlink.EVMFactoryConfig) {
			// Create ChainStores that always sign with 1337
			fc.GenChainStore = func(ks core.Keystore, i *big.Int) keys.ChainStore {
				return keys.NewChainStore(ks, big.NewInt(1337))
			}
			fc.GenEthClient = func(i *big.Int) client.Client {
				ethClient, ok := clients[i.Uint64()]
				if !ok {
					t.Fatal("no backend for chainID", i)
				}
				return ethClient
			}
		},
		Logger:                   lggr,
		ExternalInitiatorManager: nil,
		CloseLogger:              lggr.Sync,
		UnrestrictedHTTPClient:   &http.Client{},
		RestrictedHTTPClient:     &http.Client{},
		AuditLogger:              audit.NoopLogger,
	})
	require.NoError(t, err)
	keys := CreateKeys(t, app, chains, solchains)

	// JD

	setupJD(t, app)
	return &Node{
		App: app,
		Chains: slices.Concat(
			maps.Keys(chains),
			maps.Keys(solchains),
		),
		Keys:       keys,
		Addr:       net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: port},
		IsBoostrap: bootstrap,
	}
}

type Keys struct {
	PeerID        p2pkey.PeerID
	CSA           csakey.KeyV2
	WorkflowKey   workflowkey.Key
	Transmitters  map[uint64]string // chainSelector => address
	OCRKeyBundles map[chaintype.ChainType]ocr2key.KeyBundle
}

func CreateKeys(t *testing.T,
	app chainlink.Application,
	chains map[uint64]deployment.Chain,
	solchains map[uint64]deployment.SolChain,
) Keys {
	ctx := tests.Context(t)
	_, err := app.GetKeyStore().P2P().Create(ctx)
	require.NoError(t, err)

	err = app.GetKeyStore().CSA().EnsureKey(ctx)
	require.NoError(t, err)
	csaKey, err := keystore.GetDefault(ctx, app.GetKeyStore().CSA())
	require.NoError(t, err)

	p2pIDs, err := app.GetKeyStore().P2P().GetAll()
	require.NoError(t, err)
	require.Len(t, p2pIDs, 1)
	peerID := p2pIDs[0].PeerID()
	// create a transmitter for each chain
	transmitters := make(map[uint64]string)
	keybundles := make(map[chaintype.ChainType]ocr2key.KeyBundle)
	for _, chain := range chains {
		family, err := chainsel.GetSelectorFamily(chain.Selector)
		require.NoError(t, err)

		var ctype chaintype.ChainType
		switch family {
		case chainsel.FamilyEVM:
			ctype = chaintype.EVM
		case chainsel.FamilySolana:
			ctype = chaintype.Solana
		case chainsel.FamilyStarknet:
			ctype = chaintype.StarkNet
		case chainsel.FamilyCosmos:
			ctype = chaintype.Cosmos
		case chainsel.FamilyAptos:
			ctype = chaintype.Aptos
		default:
			panic(fmt.Sprintf("Unsupported chain family %v", family))
		}

		err = app.GetKeyStore().OCR2().EnsureKeys(ctx, ctype)
		require.NoError(t, err)
		keys, err := app.GetKeyStore().OCR2().GetAllOfType(ctype)
		require.NoError(t, err)
		require.Len(t, keys, 1)
		keybundle := keys[0]

		keybundles[ctype] = keybundle

		switch family {
		case chainsel.FamilyEVM:
			evmChainID, err := chainsel.ChainIdFromSelector(chain.Selector)
			require.NoError(t, err)

			cid := new(big.Int).SetUint64(evmChainID)
			addrs, err2 := app.GetKeyStore().Eth().EnabledAddressesForChain(ctx, cid)
			require.NoError(t, err2)
			var transmitter common.Address
			if len(addrs) == 1 {
				// just fund the address
				transmitter = addrs[0]
			} else {
				// create key and fund it
				_, err3 := app.GetKeyStore().Eth().Create(ctx, cid)
				require.NoError(t, err3, "failed to create key for chain", evmChainID)
				sendingKeys, err3 := app.GetKeyStore().Eth().EnabledAddressesForChain(ctx, cid)
				require.NoError(t, err3)
				require.Len(t, sendingKeys, 1)
				transmitter = sendingKeys[0]
			}
			transmitters[chain.Selector] = transmitter.String()

			backend := chain.Client.(*Backend).Sim
			fundAddress(t, chain.DeployerKey, transmitter, assets.Ether(1000).ToInt(), backend)
			// need to look more into it, but it seems like with sim chains nodes are sending txs with 0x from address
			fundAddress(t, chain.DeployerKey, common.Address{}, assets.Ether(1000).ToInt(), backend)
		case chainsel.FamilyAptos:
			err = app.GetKeyStore().Aptos().EnsureKey(ctx)
			require.NoError(t, err, "failed to create key for aptos")

			keys, err := app.GetKeyStore().Aptos().GetAll()
			require.NoError(t, err)
			require.Len(t, keys, 1)

			transmitter := keys[0]
			transmitters[chain.Selector] = transmitter.ID()

			// TODO: funding
		case chainsel.FamilyStarknet:
			err = app.GetKeyStore().StarkNet().EnsureKey(ctx)
			require.NoError(t, err, "failed to create key for starknet")

			keys, err := app.GetKeyStore().StarkNet().GetAll()
			require.NoError(t, err)
			require.Len(t, keys, 1)

			transmitter := keys[0]
			transmitters[chain.Selector] = transmitter.ID()

			// TODO: funding
		default:
			// TODO: other transmission keys unsupported for now
		}
	}

	for chain := range solchains {
		ctype := chaintype.Solana
		err = app.GetKeyStore().OCR2().EnsureKeys(ctx, ctype)
		require.NoError(t, err)
		keys, err := app.GetKeyStore().OCR2().GetAllOfType(ctype)
		require.NoError(t, err)
		require.Len(t, keys, 1)
		keybundle := keys[0]

		keybundles[ctype] = keybundle

		err = app.GetKeyStore().Solana().EnsureKey(ctx)
		require.NoError(t, err, "failed to create key for solana")

		solkeys, err := app.GetKeyStore().Solana().GetAll()
		require.NoError(t, err)
		require.Len(t, solkeys, 1)

		transmitter := solkeys[0]
		transmitters[chain] = transmitter.ID()

		// TODO: funding
	}

	return Keys{
		PeerID:        peerID,
		CSA:           csaKey,
		Transmitters:  transmitters,
		OCRKeyBundles: keybundles,
	}
}

func createConfigV2Chain(chainID uint64) *v2toml.EVMConfig {
	chainIDBig := evmutils.NewI(int64(chainID))
	chain := v2toml.Defaults(chainIDBig)
	chain.GasEstimator.LimitDefault = ptr(uint64(5e6))
	chain.LogPollInterval = config.MustNewDuration(500 * time.Millisecond)
	chain.Transactions.ForwardersEnabled = ptr(false)
	chain.FinalityDepth = ptr(uint32(2))
	return &v2toml.EVMConfig{
		ChainID: chainIDBig,
		Enabled: ptr(true),
		Chain:   chain,
		Nodes:   v2toml.EVMNodes{&v2toml.Node{}},
	}
}

func createSolanaChainConfig(chainID string, chain deployment.SolChain) *solcfg.TOMLConfig {
	chainConfig := solcfg.Chain{}
	chainConfig.SetDefaults()

	url, err := config.ParseURL(chain.URL)
	if err != nil {
		panic(err)
	}

	return &solcfg.TOMLConfig{
		ChainID:   &chainID,
		Enabled:   ptr(true),
		Chain:     chainConfig,
		MultiNode: mnCfg.MultiNodeConfig{},
		Nodes: []*solcfg.Node{{
			Name:     ptr("primary"),
			URL:      url,
			SendOnly: false,
		}},
	}
}

func ptr[T any](v T) *T { return &v }

func setupJD(t *testing.T, app chainlink.Application) {
	secret := randomBytes32(t)
	pkey, err := crypto.PublicKeyFromHex(hex.EncodeToString(secret))
	require.NoError(t, err)
	m := feeds2.RegisterManagerParams{
		Name:      "In memory env test",
		URI:       "http://dev.null:8080",
		PublicKey: *pkey,
	}
	f := app.GetFeedsService()
	connManager := feedsMocks.NewConnectionsManager(t)
	connManager.On("Connect", mock.Anything).Maybe()
	connManager.On("GetClient", mock.Anything).Maybe().Return(noopFeedsClient{}, nil)
	connManager.On("Close").Maybe().Return()
	connManager.On("IsConnected", mock.Anything).Maybe().Return(true)
	f.Unsafe_SetConnectionsManager(connManager)

	_, err = f.RegisterManager(testutils.Context(t), m)
	require.NoError(t, err)
}

func randomBytes32(t *testing.T) []byte {
	t.Helper()
	b := make([]byte, 32)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return b
}

type noopFeedsClient struct{}

func (n noopFeedsClient) ApprovedJob(context.Context, *pb.ApprovedJobRequest) (*pb.ApprovedJobResponse, error) {
	return &pb.ApprovedJobResponse{}, nil
}

func (n noopFeedsClient) Healthcheck(context.Context, *pb.HealthcheckRequest) (*pb.HealthcheckResponse, error) {
	return &pb.HealthcheckResponse{}, nil
}

func (n noopFeedsClient) UpdateNode(context.Context, *pb.UpdateNodeRequest) (*pb.UpdateNodeResponse, error) {
	return &pb.UpdateNodeResponse{}, nil
}

func (n noopFeedsClient) RejectedJob(context.Context, *pb.RejectedJobRequest) (*pb.RejectedJobResponse, error) {
	return &pb.RejectedJobResponse{}, nil
}

func (n noopFeedsClient) CancelledJob(context.Context, *pb.CancelledJobRequest) (*pb.CancelledJobResponse, error) {
	return &pb.CancelledJobResponse{}, nil
}

// Copyright 2014 The go-ethereum Authors
// This file is part of the go-haaereum library.
//
// The go-haaereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-haaereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-haaereum library. If not, see <http://www.gnu.org/licenses/>.

// Package haa implements the haachain protocol.
package haa

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/haachain/go-haachain/accounts"
	"github.com/haachain/go-haachain/common"
	"github.com/haachain/go-haachain/common/hexutil"
	"github.com/haachain/go-haachain/consensus"
	"github.com/haachain/go-haachain/consensus/clique"
	"github.com/haachain/go-haachain/consensus/ethash"
	"github.com/haachain/go-haachain/core"
	"github.com/haachain/go-haachain/core/bloombits"
	"github.com/haachain/go-haachain/core/types"
	"github.com/haachain/go-haachain/core/vm"
	"github.com/haachain/go-haachain/haa/downloader"
	"github.com/haachain/go-haachain/haa/filters"
	"github.com/haachain/go-haachain/haa/gasprice"
	"github.com/haachain/go-haachain/haadb"
	"github.com/haachain/go-haachain/event"
	"github.com/haachain/go-haachain/internal/ethapi"
	"github.com/haachain/go-haachain/log"
	"github.com/haachain/go-haachain/miner"
	"github.com/haachain/go-haachain/node"
	"github.com/haachain/go-haachain/p2p"
	"github.com/haachain/go-haachain/params"
	"github.com/haachain/go-haachain/rlp"
	"github.com/haachain/go-haachain/rpc"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// haachain implements the haachain full node service.
type haachain struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan  chan bool    // Channel for shutting down the haaereum
	stopDbUpgrade func() error // stop chain db sequential key upgrade

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb haadb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend *haaApiBackend

	miner     *miner.Miner
	gasPrice  *big.Int
	haaerbase common.Address

	networkId     uint64
	netRPCService *ethapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and haaerbase)
}

func (s *haachain) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// New creates a new haachain object (including the
// initialisation of the common haachain object)
func New(ctx *node.ServiceContext, config *Config) (*haachain, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run haa.haachain in light sync mode, use les.Lighthaachain")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	stopDbUpgrade := upgradeDeduplicateData(chainDb)
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	haa := &haachain{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, &config.haaash, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		stopDbUpgrade:  stopDbUpgrade,
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		haaerbase:      config.haaerbase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, params.BloomBitsBlocks),
	}

	log.Info("Initialising haachain protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := core.GetBlockChainVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run ghaa upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		core.WriteBlockChainVersion(chainDb, core.BlockChainVersion)
	}
	var (
		vmConfig    = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		cacheConfig = &core.CacheConfig{Disabled: config.NoPruning, TrieNodeLimit: config.TrieCache, TrieTimeLimit: config.TrieTimeout}
	)
	haa.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, haa.chainConfig, haa.engine, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		haa.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	haa.bloomIndexer.Start(haa.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	haa.txPool = core.NewTxPool(config.TxPool, haa.chainConfig, haa.blockchain)

	if haa.protocolManager, err = NewProtocolManager(haa.chainConfig, config.SyncMode, config.NetworkId, haa.eventMux, haa.txPool, haa.engine, haa.blockchain, chainDb); err != nil {
		return nil, err
	}
	haa.miner = miner.New(haa, haa.chainConfig, haa.EventMux(), haa.engine)
	haa.miner.SetExtra(makeExtraData(config.ExtraData))

	haa.ApiBackend = &haaApiBackend{haa, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	haa.ApiBackend.gpo = gasprice.NewOracle(haa.ApiBackend, gpoParams)

	return haa, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"ghaa",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (haadb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*haadb.LDBDatabase); ok {
		db.Meter("haa/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an haachain service
func CreateConsensusEngine(ctx *node.ServiceContext, config *ethash.Config, chainConfig *params.ChainConfig, db haadb.Database) consensus.Engine {
	// If proof-of-authority is requested, set it up
	if chainConfig.Clique != nil {
		return clique.New(chainConfig.Clique, db)
	}
	// Otherwise assume proof-of-work
	switch {
	case config.PowMode == ethash.ModeFake:
		log.Warn("haaash used in fake mode")
		return ethash.NewFaker()
	case config.PowMode == ethash.ModeTest:
		log.Warn("haaash used in test mode")
		return ethash.NewTester()
	case config.PowMode == ethash.ModeShared:
		log.Warn("haaash used in shared mode")
		return ethash.NewShared()
	default:
		engine := ethash.New(ethash.Config{
			CacheDir:       ctx.ResolvePath(config.CacheDir),
			CachesInMem:    config.CachesInMem,
			CachesOnDisk:   config.CachesOnDisk,
			DatasetDir:     config.DatasetDir,
			DatasetsInMem:  config.DatasetsInMem,
			DatasetsOnDisk: config.DatasetsOnDisk,
		})
		engine.SetThreads(-1) // Disable CPU mining
		return engine
	}
}

// APIs returns the collection of RPC services the haaereum package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *haachain) APIs() []rpc.API {
	apis := ethapi.GetAPIs(s.ApiBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "haa",
			Version:   "1.0",
			Service:   NewPublichaachainAPI(s),
			Public:    true,
		}, {
			Namespace: "haa",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "haa",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "haa",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *haachain) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *haachain) haaerbase() (eb common.Address, err error) {
	s.lock.RLock()
	haaerbase := s.haaerbase
	s.lock.RUnlock()

	if haaerbase != (common.Address{}) {
		return haaerbase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			haaerbase := accounts[0].Address

			s.lock.Lock()
			s.haaerbase = haaerbase
			s.lock.Unlock()

			log.Info("haaerbase automatically configured", "address", haaerbase)
			return haaerbase, nil
		}
	}
	return common.Address{}, fmt.Errorf("haaerbase must be explicitly specified")
}

// set in js console via admin interface or wrapper from cli flags
func (self *haachain) Sethaaerbase(haaerbase common.Address) {
	self.lock.Lock()
	self.haaerbase = haaerbase
	self.lock.Unlock()

	self.miner.Sethaaerbase(haaerbase)
}

func (s *haachain) StartMining(local bool) error {
	eb, err := s.haaerbase()
	if err != nil {
		log.Error("Cannot start mining without haaerbase", "err", err)
		return fmt.Errorf("haaerbase missing: %v", err)
	}
	if clique, ok := s.engine.(*clique.Clique); ok {
		wallet, err := s.accountManager.Find(accounts.Account{Address: eb})
		if wallet == nil || err != nil {
			log.Error("haaerbase account unavailable locally", "err", err)
			return fmt.Errorf("signer missing: %v", err)
		}
		clique.Authorize(eb, wallet.SignHash)
	}
	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so noone will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}
	go s.miner.Start(eb)
	return nil
}

func (s *haachain) StopMining()         { s.miner.Stop() }
func (s *haachain) IsMining() bool      { return s.miner.Mining() }
func (s *haachain) Miner() *miner.Miner { return s.miner }

func (s *haachain) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *haachain) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *haachain) TxPool() *core.TxPool               { return s.txPool }
func (s *haachain) EventMux() *event.TypeMux           { return s.eventMux }
func (s *haachain) Engine() consensus.Engine           { return s.engine }
func (s *haachain) ChainDb() haadb.Database            { return s.chainDb }
func (s *haachain) IsListening() bool                  { return true } // Always listening
func (s *haachain) haaVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *haachain) NetVersion() uint64                 { return s.networkId }
func (s *haachain) Downloader() *downloader.Downloader { return s.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *haachain) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	}
	return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// haachain protocol implementation.
func (s *haachain) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = ethapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if s.config.LightServ > 0 {
		if s.config.LightPeers >= srvr.MaxPeers {
			return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", s.config.LightPeers, srvr.MaxPeers)
		}
		maxPeers -= s.config.LightPeers
	}
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// haachain protocol.
func (s *haachain) Stop() error {
	if s.stopDbUpgrade != nil {
		s.stopDbUpgrade()
	}
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}

// Copyright 2016 The go-ethereum Authors
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

// Package les implements the Light haachain Subprotocol.
package les

import (
	"fmt"
	"sync"
	"time"

	"github.com/haachain/go-haachain/accounts"
	"github.com/haachain/go-haachain/common"
	"github.com/haachain/go-haachain/common/hexutil"
	"github.com/haachain/go-haachain/consensus"
	"github.com/haachain/go-haachain/core"
	"github.com/haachain/go-haachain/core/bloombits"
	"github.com/haachain/go-haachain/core/types"
	"github.com/haachain/go-haachain/haa"
	"github.com/haachain/go-haachain/haa/downloader"
	"github.com/haachain/go-haachain/haa/filters"
	"github.com/haachain/go-haachain/haa/gasprice"
	"github.com/haachain/go-haachain/haadb"
	"github.com/haachain/go-haachain/event"
	"github.com/haachain/go-haachain/internal/ethapi"
	"github.com/haachain/go-haachain/light"
	"github.com/haachain/go-haachain/log"
	"github.com/haachain/go-haachain/node"
	"github.com/haachain/go-haachain/p2p"
	"github.com/haachain/go-haachain/p2p/discv5"
	"github.com/haachain/go-haachain/params"
	rpc "github.com/haachain/go-haachain/rpc"
)

type Lighthaachain struct {
	config *haa.Config

	odr         *LesOdr
	relay       *LesTxRelay
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool
	// Handlers
	peers           *peerSet
	txPool          *light.TxPool
	blockchain      *light.LightChain
	protocolManager *ProtocolManager
	serverPool      *serverPool
	reqDist         *requestDistributor
	retriever       *retrieveManager
	// DB interfaces
	chainDb haadb.Database // Block chain database

	bloomRequests                              chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer, chtIndexer, bloomTrieIndexer *core.ChainIndexer

	ApiBackend *LesApiBackend

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *ethapi.PublicNetAPI

	wg sync.WaitGroup
}

func New(ctx *node.ServiceContext, config *haa.Config) (*Lighthaachain, error) {
	chainDb, err := haa.CreateDB(ctx, config, "lightchaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newPeerSet()
	quitSync := make(chan struct{})

	lhaa := &Lighthaachain{
		config:           config,
		chainConfig:      chainConfig,
		chainDb:          chainDb,
		eventMux:         ctx.EventMux,
		peers:            peers,
		reqDist:          newRequestDistributor(peers, quitSync),
		accountManager:   ctx.AccountManager,
		engine:           haa.CreateConsensusEngine(ctx, &config.haaash, chainConfig, chainDb),
		shutdownChan:     make(chan bool),
		networkId:        config.NetworkId,
		bloomRequests:    make(chan chan *bloombits.Retrieval),
		bloomIndexer:     haa.NewBloomIndexer(chainDb, light.BloomTrieFrequency),
		chtIndexer:       light.NewChtIndexer(chainDb, true),
		bloomTrieIndexer: light.NewBloomTrieIndexer(chainDb, true),
	}

	lhaa.relay = NewLesTxRelay(peers, lhaa.reqDist)
	lhaa.serverPool = newServerPool(chainDb, quitSync, &lhaa.wg)
	lhaa.retriever = newRetrieveManager(peers, lhaa.reqDist, lhaa.serverPool)
	lhaa.odr = NewLesOdr(chainDb, lhaa.chtIndexer, lhaa.bloomTrieIndexer, lhaa.bloomIndexer, lhaa.retriever)
	if lhaa.blockchain, err = light.NewLightChain(lhaa.odr, lhaa.chainConfig, lhaa.engine); err != nil {
		return nil, err
	}
	lhaa.bloomIndexer.Start(lhaa.blockchain)
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		lhaa.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	lhaa.txPool = light.NewTxPool(lhaa.chainConfig, lhaa.blockchain, lhaa.relay)
	if lhaa.protocolManager, err = NewProtocolManager(lhaa.chainConfig, true, ClientProtocolVersions, config.NetworkId, lhaa.eventMux, lhaa.engine, lhaa.peers, lhaa.blockchain, nil, chainDb, lhaa.odr, lhaa.relay, quitSync, &lhaa.wg); err != nil {
		return nil, err
	}
	lhaa.ApiBackend = &LesApiBackend{lhaa, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	lhaa.ApiBackend.gpo = gasprice.NewOracle(lhaa.ApiBackend, gpoParams)
	return lhaa, nil
}

func lesTopic(genesisHash common.Hash, protocolVersion uint) discv5.Topic {
	var name string
	switch protocolVersion {
	case lpv1:
		name = "LES"
	case lpv2:
		name = "LES2"
	default:
		panic(nil)
	}
	return discv5.Topic(name + "@" + common.Bytes2Hex(genesisHash.Bytes()[0:8]))
}

type LightDummyAPI struct{}

// haaerbase is the address that mining rewards will be send to
func (s *LightDummyAPI) haaerbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Coinbase is the address that mining rewards will be send to (alias for haaerbase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the haaereum package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Lighthaachain) APIs() []rpc.API {
	return append(ethapi.GetAPIs(s.ApiBackend), []rpc.API{
		{
			Namespace: "haa",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "haa",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "haa",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *Lighthaachain) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Lighthaachain) BlockChain() *light.LightChain      { return s.blockchain }
func (s *Lighthaachain) TxPool() *light.TxPool              { return s.txPool }
func (s *Lighthaachain) Engine() consensus.Engine           { return s.engine }
func (s *Lighthaachain) LesVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Lighthaachain) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *Lighthaachain) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Lighthaachain) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

// Start implements node.Service, starting all internal goroutines needed by the
// haachain protocol implementation.
func (s *Lighthaachain) Start(srvr *p2p.Server) error {
	s.startBloomHandlers()
	log.Warn("Light client mode is an experimental feature")
	s.netRPCService = ethapi.NewPublicNetAPI(srvr, s.networkId)
	// clients are searching for the first advertised protocol in the list
	protocolVersion := AdvertiseProtocolVersions[0]
	s.serverPool.start(srvr, lesTopic(s.blockchain.Genesis().Hash(), protocolVersion))
	s.protocolManager.Start(s.config.LightPeers)
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// haachain protocol.
func (s *Lighthaachain) Stop() error {
	s.odr.Stop()
	if s.bloomIndexer != nil {
		s.bloomIndexer.Close()
	}
	if s.chtIndexer != nil {
		s.chtIndexer.Close()
	}
	if s.bloomTrieIndexer != nil {
		s.bloomTrieIndexer.Close()
	}
	s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()

	s.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}

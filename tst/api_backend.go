// Copyright 2015 The go-ethereum Authors
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

package haa

import (
	"context"
	"math/big"

	"github.com/haachain/go-haachain/accounts"
	"github.com/haachain/go-haachain/common"
	"github.com/haachain/go-haachain/common/math"
	"github.com/haachain/go-haachain/core"
	"github.com/haachain/go-haachain/core/bloombits"
	"github.com/haachain/go-haachain/core/state"
	"github.com/haachain/go-haachain/core/types"
	"github.com/haachain/go-haachain/core/vm"
	"github.com/haachain/go-haachain/haa/downloader"
	"github.com/haachain/go-haachain/haa/gasprice"
	"github.com/haachain/go-haachain/haadb"
	"github.com/haachain/go-haachain/event"
	"github.com/haachain/go-haachain/params"
	"github.com/haachain/go-haachain/rpc"
)

// haaApiBackend implements ethapi.Backend for full nodes
type haaApiBackend struct {
	haa *haachain
	gpo *gasprice.Oracle
}

func (b *haaApiBackend) ChainConfig() *params.ChainConfig {
	return b.haa.chainConfig
}

func (b *haaApiBackend) CurrentBlock() *types.Block {
	return b.haa.blockchain.CurrentBlock()
}

func (b *haaApiBackend) SetHead(number uint64) {
	b.haa.protocolManager.downloader.Cancel()
	b.haa.blockchain.SetHead(number)
}

func (b *haaApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.haa.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.haa.blockchain.CurrentBlock().Header(), nil
	}
	return b.haa.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *haaApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.haa.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.haa.blockchain.CurrentBlock(), nil
	}
	return b.haa.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *haaApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.haa.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.haa.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *haaApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.haa.blockchain.GetBlockByHash(blockHash), nil
}

func (b *haaApiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	return core.GetBlockReceipts(b.haa.chainDb, blockHash, core.GetBlockNumber(b.haa.chainDb, blockHash)), nil
}

func (b *haaApiBackend) GetLogs(ctx context.Context, blockHash common.Hash) ([][]*types.Log, error) {
	receipts := core.GetBlockReceipts(b.haa.chainDb, blockHash, core.GetBlockNumber(b.haa.chainDb, blockHash))
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *haaApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.haa.blockchain.GetTdByHash(blockHash)
}

func (b *haaApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.haa.BlockChain(), nil)
	return vm.NewEVM(context, state, b.haa.chainConfig, vmCfg), vmError, nil
}

func (b *haaApiBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.haa.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *haaApiBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.haa.BlockChain().SubscribeChainEvent(ch)
}

func (b *haaApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.haa.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *haaApiBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.haa.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *haaApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.haa.BlockChain().SubscribeLogsEvent(ch)
}

func (b *haaApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.haa.txPool.AddLocal(signedTx)
}

func (b *haaApiBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.haa.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *haaApiBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.haa.txPool.Get(hash)
}

func (b *haaApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.haa.txPool.State().GetNonce(addr), nil
}

func (b *haaApiBackend) Stats() (pending int, queued int) {
	return b.haa.txPool.Stats()
}

func (b *haaApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.haa.TxPool().Content()
}

func (b *haaApiBackend) SubscribeTxPreEvent(ch chan<- core.TxPreEvent) event.Subscription {
	return b.haa.TxPool().SubscribeTxPreEvent(ch)
}

func (b *haaApiBackend) Downloader() *downloader.Downloader {
	return b.haa.Downloader()
}

func (b *haaApiBackend) ProtocolVersion() int {
	return b.haa.haaVersion()
}

func (b *haaApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *haaApiBackend) ChainDb() haadb.Database {
	return b.haa.ChainDb()
}

func (b *haaApiBackend) EventMux() *event.TypeMux {
	return b.haa.EventMux()
}

func (b *haaApiBackend) AccountManager() *accounts.Manager {
	return b.haa.AccountManager()
}

func (b *haaApiBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.haa.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *haaApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.haa.bloomRequests)
	}
}

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

package haaclient

import "github.com/haachain/go-haachain"

// Verify that Client implements the haaereum interfaces.
var (
	_ = haaereum.ChainReader(&Client{})
	_ = haaereum.TransactionReader(&Client{})
	_ = haaereum.ChainStateReader(&Client{})
	_ = haaereum.ChainSyncReader(&Client{})
	_ = haaereum.ContractCaller(&Client{})
	_ = haaereum.GasEstimator(&Client{})
	_ = haaereum.GasPricer(&Client{})
	_ = haaereum.LogFilterer(&Client{})
	_ = haaereum.PendingStateReader(&Client{})
	// _ = haaereum.PendingStateEventer(&Client{})
	_ = haaereum.PendingContractCaller(&Client{})
)

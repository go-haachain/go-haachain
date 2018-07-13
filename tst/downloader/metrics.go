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

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/haachain/go-haachain/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("haa/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("haa/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("haa/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("haa/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("haa/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("haa/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("haa/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("haa/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("haa/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("haa/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("haa/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("haa/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("haa/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("haa/downloader/states/drop", nil)
)

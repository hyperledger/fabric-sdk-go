/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"regexp"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// BlockReg contains the data for a block registration
type BlockReg struct {
	Filter  fab.BlockFilter
	Eventch chan<- *fab.BlockEvent
}

// FilteredBlockReg contains the data for a filtered block registration
type FilteredBlockReg struct {
	Eventch chan<- *fab.FilteredBlockEvent
}

// ChaincodeReg contains the data for a chaincode registration
type ChaincodeReg struct {
	ChaincodeID string
	EventFilter string
	EventRegExp *regexp.Regexp
	Eventch     chan<- *fab.CCEvent
}

// TxStatusReg contains the data for a transaction status registration
type TxStatusReg struct {
	TxID    string
	Eventch chan<- *fab.TxStatusEvent
}

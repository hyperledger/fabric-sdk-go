/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"regexp"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// BlockReg contains the data for a block registration
type BlockReg struct {
	Filter  apifabclient.BlockFilter
	Eventch chan<- *apifabclient.BlockEvent
}

// FilteredBlockReg contains the data for a filtered block registration
type FilteredBlockReg struct {
	Eventch chan<- *apifabclient.FilteredBlockEvent
}

// ChaincodeReg contains the data for a chaincode registration
type ChaincodeReg struct {
	ChaincodeID string
	EventFilter string
	EventRegExp *regexp.Regexp
	Eventch     chan<- *apifabclient.CCEvent
}

// TxStatusReg contains the data for a transaction status registration
type TxStatusReg struct {
	TxID    string
	Eventch chan<- *apifabclient.TxStatusEvent
}

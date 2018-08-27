/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"fmt"
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

type snapshot struct {
	lastBlockReceived          uint64
	blockRegistrations         []*BlockReg
	filteredBlockRegistrations []*FilteredBlockReg
	ccRegistrations            []*ChaincodeReg
	txStatusRegistrations      []*TxStatusReg
}

func (s *snapshot) LastBlockReceived() uint64 {
	return s.lastBlockReceived
}

func (s *snapshot) BlockRegistrations() []fab.Registration {
	return fromBlockReg(s.blockRegistrations)
}

func (s *snapshot) FilteredBlockRegistrations() []fab.Registration {
	return fromFBlockReg(s.filteredBlockRegistrations)
}

func (s *snapshot) CCRegistrations() []fab.Registration {
	return fromCCReg(s.ccRegistrations)
}

func (s *snapshot) TxStatusRegistrations() []fab.Registration {
	return fromTxReg(s.txStatusRegistrations)
}

func (s *snapshot) String() string {
	var ccReg []string
	for _, reg := range s.ccRegistrations {
		ccReg = append(ccReg, fmt.Sprintf("{ID: %s, Event: %s}", reg.ChaincodeID, reg.EventFilter))
	}

	var txReg []string
	for _, reg := range s.txStatusRegistrations {
		txReg = append(txReg, fmt.Sprintf("{TxID: %s}", reg.TxID))
	}

	return fmt.Sprintf("Last Block: %d, Block Reg's: %d, Filtered Block Reg's: %d, CC Reg's: %s, TxStatus Reg's: %s",
		s.lastBlockReceived, len(s.blockRegistrations), len(s.filteredBlockRegistrations), ccReg, txReg)
}

// Close closes all event registrations
func (s *snapshot) Close() {
	for _, reg := range s.blockRegistrations {
		close(reg.Eventch)
	}
	for _, reg := range s.filteredBlockRegistrations {
		close(reg.Eventch)
	}
	for _, reg := range s.ccRegistrations {
		close(reg.Eventch)
	}
	for _, reg := range s.txStatusRegistrations {
		close(reg.Eventch)
	}
}

func fromBlockReg(bRegistrations []*BlockReg) []fab.Registration {
	var registrations []fab.Registration
	for _, reg := range bRegistrations {
		registrations = append(registrations, reg)
	}
	return registrations
}

func fromFBlockReg(bRegistrations []*FilteredBlockReg) []fab.Registration {
	var registrations []fab.Registration
	for _, reg := range bRegistrations {
		registrations = append(registrations, reg)
	}
	return registrations
}

func fromCCReg(bRegistrations []*ChaincodeReg) []fab.Registration {
	var registrations []fab.Registration
	for _, reg := range bRegistrations {
		registrations = append(registrations, reg)
	}
	return registrations
}

func fromTxReg(bRegistrations []*TxStatusReg) []fab.Registration {
	var registrations []fab.Registration
	for _, reg := range bRegistrations {
		registrations = append(registrations, reg)
	}
	return registrations
}

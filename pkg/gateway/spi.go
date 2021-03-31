/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

// This contains the service provider interface (SPI) which provides the mechanism
// for implementing alternative gateway strategies, wallets, etc.
// This is currently experimental and will be implemented in future user stories

// WalletStore is the interface for implementations that provide backing storage for identities in a wallet.
// To create create a new backing store, implement all the methods defined in this interface and provide
// a factory method that wraps an instance of this in a new Wallet object. E.g:
//   func NewMyWallet() *Wallet {
//	   store := &myWalletStore{ }
//	   return &Wallet{store}
//   }
type WalletStore interface {
	Put(label string, stream []byte) error
	Get(label string) ([]byte, error)
	List() ([]string, error)
	Exists(label string) bool
	Remove(label string) error
}

// CheckpointStore is the interface for implementations that need to persist current block and transactions to
// enable event listening to be resumed following an application failure.
// To create create a new checkpòinter, implement all the methods defined in this interface and provide
// a factory method that wraps an instance of this in a new Checkpoint object. E.g:
//   func NewCheckpoint() *Checkpoint {
//	   store := &myCheckpoint{ }
//	   return &Checkpoint{store}
//   }
type CheckpointStore interface {
	GetBlockNumber() uint64
	SetBlockNumber(blockNumber uint64)
	GetTransactionIds() []string
	AddTransactionId(transactionId string)
}


/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	mspProvider "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

type gatewayOptions struct {
	Identity      mspProvider.SigningIdentity
	User          string
	CommitHandler CommitHandlerFactory
	Discovery     bool
}

// Option functional arguments can be supplied when connecting to the gateway.
type Option = func(Gateway, *gatewayOptions) error

// ConfigOption specifies the gateway configuration source.
type ConfigOption = func(Gateway, *gatewayOptions) error

// IdentityOption specifies the user identity under which all transactions are performed for this gateway instance.
type IdentityOption = func(Gateway, *gatewayOptions) error

// Gateway is the entry point to a Fabric network
type Gateway interface {
	GetNetwork(string) (Network, error)
	Close()
	getSdk() *fabsdk.FabricSDK
	getOrg() string
	getPeersForOrg(string) ([]string, error)
}

// A Network object represents the set of peers in a Fabric network (channel).
// Applications should get a Network instance from a Gateway using the GetNetwork method.
type Network interface {
	GetContract(string) Contract
	GetName() string
}

// A Contract object represents a smart contract instance in a network.
// Applications should get a Contract instance from a Network using the GetContract method
type Contract interface {
	GetName() string
	EvaluateTransaction(string, ...string) ([]byte, error)
	SubmitTransaction(string, ...string) ([]byte, error)
	CreateTransaction(string, ...TransactionOption) (Transaction, error)
}

// type transactionOptions struct {
// 	Transient      map[string][]byte
// 	EndorsingPeers []string
// }

// TransactionOption functional arguments can be supplied when creating a transaction object
type TransactionOption = func(*transaction) error

// A Transaction represents a specific invocation of a transaction function, and provides
// flexibility over how that transaction is invoked. Applications should
// obtain instances of this class from a Contract using the
// Contract.CreateTransaction method.
//
// Instances of this class are stateful. A new instance <strong>must</strong>
// be created for each transaction invocation.
type Transaction interface {
	Evaluate(...string) ([]byte, error)
	Submit(...string) ([]byte, error)
}

// A Wallet stores identity information used to connect to a Hyperledger Fabric network.
// Instances are created using factory methods on the implementing objects.
type Wallet interface {
	Put(label string, id IdentityType) error
	Get(label string) (IdentityType, error)
	Remove(label string) error
	Exists(label string) bool
	List() []string
}

// SPI...

// CommitHandlerFactory is currently unimplemented
type CommitHandlerFactory interface {
	Create(string, Network) CommitHandler
}

// CommitHandler is currently unimplemented
type CommitHandler interface {
	StartListening()
	WaitForEvents(int64)
	CancelListening()
}

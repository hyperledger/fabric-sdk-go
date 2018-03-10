/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defmsp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/identitymgr"
)

// ProviderFactory represents the default MSP provider factory.
type ProviderFactory struct {
}

// NewProviderFactory returns the default MSP provider factory.
func NewProviderFactory() *ProviderFactory {
	f := ProviderFactory{}
	return &f
}

// CreateIdentityManager returns a new default implementation of identity manager
func (f *ProviderFactory) CreateIdentityManager(org string, stateStore core.KVStore, cryptoProvider core.CryptoSuite, config core.Config) (msp.IdentityManager, error) {
	return identitymgr.New(org, stateStore, cryptoProvider, config)
}

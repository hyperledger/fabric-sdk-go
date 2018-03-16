/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defmsp"
)

// ========== MSP Provider Factory with custom user store ============= //

// CustomMSPFactory is a custom factory for tests.
type CustomMSPFactory struct {
	defaultFactory  *defmsp.ProviderFactory
	customUserStore msp.UserStore
}

// NewCustomMSPFactory creates a custom MSPFactory
func NewCustomMSPFactory(customUserStore msp.UserStore) *CustomMSPFactory {
	return &CustomMSPFactory{defaultFactory: defmsp.NewProviderFactory(), customUserStore: customUserStore}
}

// CreateUserStore creates UserStore
func (f *CustomMSPFactory) CreateUserStore(config core.Config) (msp.UserStore, error) {
	return f.customUserStore, nil
}

// CreateProvider creates an MSP provider
func (f *CustomMSPFactory) CreateProvider(config core.Config, cryptoProvider core.CryptoSuite, userStore msp.UserStore) (msp.Provider, error) {
	return f.defaultFactory.CreateProvider(config, cryptoProvider, f.customUserStore)
}

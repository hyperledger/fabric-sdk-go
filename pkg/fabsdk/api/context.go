/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicore"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// Providers represents the configured providers context
type Providers interface {
	CoreProviders
	SvcProviders
}

// CoreProviders represents the configured core providers context
type CoreProviders interface {
	CryptoSuiteProvider() apicryptosuite.CryptoSuite
	StateStoreProvider() fab.KeyValueStore
	ConfigProvider() apiconfig.Config
	SigningManager() fab.SigningManager
	FabricProvider() apicore.FabricProvider
}

// SvcProviders represents the configured service providers context
type SvcProviders interface {
	DiscoveryProvider() fab.DiscoveryProvider
	SelectionProvider() fab.SelectionProvider
}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// SDK represents the configuration context
type SDK interface {
	CryptoSuiteProvider() apicryptosuite.CryptoSuite
	StateStoreProvider() fab.KeyValueStore
	ConfigProvider() apiconfig.Config
	DiscoveryProvider() fab.DiscoveryProvider
	SelectionProvider() fab.SelectionProvider
	SigningManager() fab.SigningManager
}

// Org represents the organization context
type Org interface {
	// TODO
}

// Session primarily represents the session and identity context
type Session interface {
	Identity() fab.User
}

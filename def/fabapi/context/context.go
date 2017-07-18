/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric/bccsp"
)

// SDK represents the configuration context
type SDK interface {
	CryptoSuiteProvider() bccsp.BCCSP
	StateStoreProvider() fab.KeyValueStore
	ConfigProvider() apiconfig.Config
}

// Org represents the organization context
type Org interface {
	// TODO
}

// Session primarily represents the session and identity context
type Session interface {
	Identity() fab.User
}

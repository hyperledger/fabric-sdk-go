/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"

// Context is the context required by MSP services
type Context interface {
	core.Providers
	Providers
}

// IdentityManagerProvider provides identity management services
type IdentityManagerProvider interface {
	IdentityManager(orgName string) (IdentityManager, bool)
}

// Providers represents a provider of MSP service.
type Providers interface {
	UserStore() UserStore
	IdentityManagerProvider
}

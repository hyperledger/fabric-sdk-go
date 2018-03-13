/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import "github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"

// Context is the context required by MSP services
type Context interface {
	core.Providers
	Providers
}

// Provider provides MSP services
type Provider interface {
	UserStore() UserStore
	IdentityManager(orgName string) (IdentityManager, bool)
}

// Providers represents the MSP service providers context.
type Providers interface {
	Provider
}

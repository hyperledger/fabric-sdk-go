/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msppvdr

import (
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	mspimpl "github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/pkg/errors"
)

// MSPProvider provides the default implementation of MSP
type MSPProvider struct {
	providerContext core.Providers
	userStore       msp.UserStore
	identityManager map[string]msp.IdentityManager
}

// New creates a MSP context provider
func New(endpointConfig fab.EndpointConfig, cryptoSuite core.CryptoSuite, userStore msp.UserStore) (*MSPProvider, error) {

	identityManager := make(map[string]msp.IdentityManager)
	netConfig := endpointConfig.NetworkConfig()
	for orgName := range netConfig.Organizations {
		mgr, err := mspimpl.NewIdentityManager(orgName, userStore, cryptoSuite, endpointConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to initialize identity manager for organization: %s", orgName)
		}
		identityManager[orgName] = mgr
	}

	mspProvider := MSPProvider{
		userStore:       userStore,
		identityManager: identityManager,
	}

	return &mspProvider, nil
}

// Initialize sets the provider context
func (p *MSPProvider) Initialize(providers core.Providers) error {
	p.providerContext = providers
	return nil
}

// UserStore returns the user store used by the MSP provider
func (p *MSPProvider) UserStore() msp.UserStore {
	return p.userStore
}

// IdentityManager returns the organization's identity manager
func (p *MSPProvider) IdentityManager(orgName string) (msp.IdentityManager, bool) {
	im, ok := p.identityManager[strings.ToLower(orgName)]
	if !ok {
		return nil, false
	}
	return im, true
}

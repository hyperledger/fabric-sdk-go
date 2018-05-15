/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// LocalContext supplies the configuration for channel context client
type LocalContext struct {
	*MockContext
	localDiscovery fab.DiscoveryService
}

// LocalDiscoveryService returns the local discovery service
func (c *LocalContext) LocalDiscoveryService() fab.DiscoveryService {
	return c.localDiscovery
}

// NewMockLocalContext creates new mock local context
func NewMockLocalContext(client *MockContext, discoveryProvider fab.LocalDiscoveryProvider) *LocalContext {
	var localDiscovery fab.DiscoveryService
	if discoveryProvider != nil {
		var err error
		localDiscovery, err = discoveryProvider.CreateLocalDiscoveryService(client.Identifier().MSPID)
		if err != nil {
			panic(fmt.Sprintf("error creating local discovery service: %s", err))
		}

	}

	return &LocalContext{
		MockContext:    client,
		localDiscovery: localDiscovery,
	}
}

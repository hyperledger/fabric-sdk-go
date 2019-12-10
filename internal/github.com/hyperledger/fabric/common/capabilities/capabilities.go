/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package capabilities

import (
	cb "github.com/hyperledger/fabric-protos-go/common"
	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
	"github.com/pkg/errors"
)

var logger = flogging.MustGetLogger("common.capabilities")

// provider is the 'plugin' parameter for registry.
type provider interface {
	// HasCapability should report whether the binary supports this capability.
	HasCapability(capability string) bool

	// Type is used to make error messages more legible.
	Type() string
}

// registry is a common structure intended to be used to support specific aspects of capabilities
// such as orderer, application, and channel.
type registry struct {
	provider     provider
	capabilities map[string]*cb.Capability
}

func newRegistry(p provider, capabilities map[string]*cb.Capability) *registry {
	return &registry{
		provider:     p,
		capabilities: capabilities,
	}
}

// Supported checks that all of the required capabilities are supported by this binary.
func (r *registry) Supported() error {
	for capabilityName := range r.capabilities {
		if r.provider.HasCapability(capabilityName) {
			logger.Debugf("%s capability %s is supported and is enabled", r.provider.Type(), capabilityName)
			continue
		}

		return errors.Errorf("%s capability %s is required but not supported", r.provider.Type(), capabilityName)
	}
	return nil
}

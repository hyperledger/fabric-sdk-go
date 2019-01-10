// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricselection

import (
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
)

// SetClientProvider overrides the discovery client provider for unit tests
func SetClientProvider(provider func(ctx contextAPI.Client) (DiscoveryClient, error)) {
	clientProvider = provider
}

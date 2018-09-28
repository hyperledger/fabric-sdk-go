/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package cache

import (
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
	"github.com/pkg/errors"
)

const (
	deserializeIdentityCacheSize = 100
	validateIdentityCacheSize    = 100
	satisfiesPrincipalCacheSize  = 100
)

var mspLogger = flogging.MustGetLogger("msp")

func New(o msp.MSP) (msp.MSP, error) {
	mspLogger.Debugf("Creating Cache-MSP instance")
	if o == nil {
		return nil, errors.Errorf("Invalid passed MSP. It must be different from nil.")
	}

	theMsp := &cachedMSP{MSP: o}
	theMsp.deserializeIdentityCache = newSecondChanceCache(deserializeIdentityCacheSize)
	theMsp.satisfiesPrincipalCache = newSecondChanceCache(satisfiesPrincipalCacheSize)
	theMsp.validateIdentityCache = newSecondChanceCache(validateIdentityCacheSize)

	return theMsp, nil
}

type cachedMSP struct {
	msp.MSP

	// cache for DeserializeIdentity.
	deserializeIdentityCache *secondChanceCache

	// cache for validateIdentity
	validateIdentityCache *secondChanceCache

	// basically a map of principals=>identities=>stringified to booleans
	// specifying whether this identity satisfies this principal
	satisfiesPrincipalCache *secondChanceCache
}

type cachedIdentity struct {
	msp.Identity
	cache *cachedMSP
}

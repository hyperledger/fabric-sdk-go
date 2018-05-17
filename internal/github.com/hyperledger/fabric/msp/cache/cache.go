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
	"fmt"
	"sync"

	"github.com/golang/groupcache/lru"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
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
		return nil, fmt.Errorf("Invalid passed MSP. It must be different from nil.")
	}

	theMsp := &cachedMSP{MSP: o}
	theMsp.deserializeIdentityCache = lru.New(deserializeIdentityCacheSize)
	theMsp.satisfiesPrincipalCache = lru.New(satisfiesPrincipalCacheSize)
	theMsp.validateIdentityCache = lru.New(validateIdentityCacheSize)

	return theMsp, nil
}

type cachedMSP struct {
	msp.MSP

	// cache for DeserializeIdentity.
	deserializeIdentityCache *lru.Cache

	dicMutex sync.Mutex // synchronize access to cache

	// cache for validateIdentity
	validateIdentityCache *lru.Cache

	vicMutex sync.Mutex // synchronize access to cache

	// basically a map of principals=>identities=>stringified to booleans
	// specifying whether this identity satisfies this principal
	satisfiesPrincipalCache *lru.Cache

	spcMutex sync.Mutex // synchronize access to cache
}

type cachedIdentity struct {
	msp.Identity
	cache *cachedMSP
}

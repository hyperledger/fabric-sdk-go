/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package sessioncache

import (
	"fmt"
	"time"

	"sync"

	"encoding/hex"

	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/miekg/pkcs11"
)

var sessionCache map[string]*lazycache.Cache

var logger = flogging.MustGetLogger("bccsp_p11_sessioncache")

const (
	privateKeyFlag = true
)

// keyPairCacheKey
type KeyPairCacheKey struct {
	Mod     *pkcs11.Ctx
	Session pkcs11.SessionHandle
	SKI     []byte
	KeyType bool
}

//String return string value for config key
func (keyPairCacheKey *KeyPairCacheKey) String() string {
	return fmt.Sprintf("%x_%t", keyPairCacheKey.SKI, keyPairCacheKey.KeyType)
}

func timeTrack(start time.Time, msg string) {
	elapsed := time.Since(start)
	logger.Debugf("%s took %s", msg, elapsed)
}

func ClearAllSession(rwMtx sync.RWMutex) {

	if sessionCache != nil && len(sessionCache) > 0 {
		rwMtx.Lock()
		for _, val := range sessionCache {
			val.Close()
		}
		sessionCache = nil
		rwMtx.Unlock()
	}
}

func ClearSession(rwMtx sync.RWMutex, key string) {
	rwMtx.RLock()
	val, ok := sessionCache[key]
	rwMtx.RUnlock()
	if ok {
		rwMtx.Lock()
		val.Close()
		sessionCache[key] = nil
		rwMtx.Unlock()

	}
}

func AddSession(rwMtx sync.RWMutex, key string) {
	rwMtx.RLock()
	_, ok := sessionCache[key]
	rwMtx.RUnlock()

	if !ok {
		if sessionCache == nil {
			sessionCache = make(map[string]*lazycache.Cache)
		}
		rwMtx.Lock()
		sessionCache[key] = lazycache.New(
			"KeyPair_Resolver_Cache",
			func(key lazycache.Key) (interface{}, error) {
				return lazyref.New(
					func() (interface{}, error) {
						return getKeyPairFromSKI(key.(*KeyPairCacheKey))
					},
				), nil
			})
		rwMtx.Unlock()
	}
}

func GetKeyPairFromSessionSKI(rwMtx sync.RWMutex, keyPairCacheKey *KeyPairCacheKey) (*pkcs11.ObjectHandle, error) {
	rwMtx.RLock()
	val, ok := sessionCache[fmt.Sprintf("%d", keyPairCacheKey.Session)]
	rwMtx.RUnlock()
	if ok {
		defer timeTrack(time.Now(), fmt.Sprintf("finding  key [session: %d] [ski: %x]", keyPairCacheKey.Session, keyPairCacheKey.SKI))
		value, err := val.Get(keyPairCacheKey)
		if err != nil {
			return nil, err
		}
		lazyRef := value.(*lazyref.Reference)
		resolver, err := lazyRef.Get()
		if err != nil {
			return nil, err
		}
		return resolver.(*pkcs11.ObjectHandle), nil
	}
	return nil, fmt.Errorf("cannot find session in sessionCache")
}

func getKeyPairFromSKI(keyPairCacheKey *KeyPairCacheKey) (*pkcs11.ObjectHandle, error) {
	ktype := pkcs11.CKO_PUBLIC_KEY
	if keyPairCacheKey.KeyType == privateKeyFlag {
		ktype = pkcs11.CKO_PRIVATE_KEY
	}

	template := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, ktype),
		pkcs11.NewAttribute(pkcs11.CKA_ID, keyPairCacheKey.SKI),
	}
	if err := keyPairCacheKey.Mod.FindObjectsInit(keyPairCacheKey.Session, template); err != nil {
		return nil, err
	}

	// single session instance, assume one hit only
	objs, _, err := keyPairCacheKey.Mod.FindObjects(keyPairCacheKey.Session, 1)
	if err != nil {
		return nil, err
	}
	if err = keyPairCacheKey.Mod.FindObjectsFinal(keyPairCacheKey.Session); err != nil {
		return nil, err
	}

	if len(objs) == 0 {
		return nil, fmt.Errorf("Key not found [%s]", hex.Dump(keyPairCacheKey.SKI))
	}

	return &objs[0], nil
}

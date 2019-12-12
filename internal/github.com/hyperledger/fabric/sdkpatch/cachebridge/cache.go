/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package cachebridge

import (
	"fmt"
	"time"

	"encoding/hex"

	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"
	"github.com/miekg/pkcs11"
)

var logger = flogging.MustGetLogger("bccsp_p11_sessioncache")

var sessionCache = newSessionCache()

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

//String return string value for keyPairCacheKey
func (keyPairCacheKey *KeyPairCacheKey) String() string {
	return fmt.Sprintf("%x_%t", keyPairCacheKey.SKI, keyPairCacheKey.KeyType)
}

// SessionCacheKey
type SessionCacheKey struct {
	SessionID string
}

//String return string value for SessionCacheKey
func (SessionCacheKey *SessionCacheKey) String() string {
	return SessionCacheKey.SessionID
}

func newSessionCache() *lazycache.Cache {
	return lazycache.New(
		"Session_Resolver_Cache",
		func(key lazycache.Key) (interface{}, error) {
			return lazycache.New(
				"KeyPair_Resolver_Cache",
				func(key lazycache.Key) (interface{}, error) {
					return getKeyPairFromSKI(key.(*KeyPairCacheKey))
				}), nil
		})
}

func timeTrack(start time.Time, msg string) {
	elapsed := time.Since(start)
	logger.Debugf("%s took %s", msg, elapsed)
}

func ClearAllSession() {
	sessionCache.DeleteAll()
}

func ClearSession(key string) {
	sessionCache.Delete(&SessionCacheKey{SessionID: key})
}

func GetKeyPairFromSessionSKI(keyPairCacheKey *KeyPairCacheKey) (*pkcs11.ObjectHandle, error) {
	keyPairCache, err := sessionCache.Get(&SessionCacheKey{SessionID: fmt.Sprintf("%d", keyPairCacheKey.Session)})
	if err != nil {
		return nil, err
	}
	if keyPairCache != nil {
		val := keyPairCache.(*lazycache.Cache)
		defer timeTrack(time.Now(), fmt.Sprintf("finding  key [session: %d] [ski: %x]", keyPairCacheKey.Session, keyPairCacheKey.SKI))
		value, err := val.Get(keyPairCacheKey)
		if err != nil {
			return nil, err
		}
		return value.(*pkcs11.ObjectHandle), nil
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

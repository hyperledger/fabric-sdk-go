/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cryptosuite

import (
	"sync/atomic"

	"errors"

	"sync"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/factory"
	cryptosuiteimpl "github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

var initOnce sync.Once
var defaultCryptoSuite apicryptosuite.CryptoSuite
var initialized int32

func initSuite(defaultSuite apicryptosuite.CryptoSuite) error {
	if defaultSuite == nil {
		return errors.New("attempting to set invalid default suite")
	}
	initOnce.Do(func() {
		defaultCryptoSuite = defaultSuite
		atomic.StoreInt32(&initialized, 1)
	})
	return nil
}

//GetDefault returns default apicryptosuite
func GetDefault() apicryptosuite.CryptoSuite {
	if atomic.LoadInt32(&initialized) > 0 {
		return defaultCryptoSuite
	}
	//Set default suite
	logger.Info("No default cryptosuite found, using bccsp factory default implementation")
	initSuite(cryptosuiteimpl.GetSuite(factory.GetDefault()))
	return defaultCryptoSuite
}

//SetDefault sets default suite if one is not already set or created
//Make sure you set default suite before very first call to GetDefault(),
//otherwise this function will return an error
func SetDefault(newDefaultSuite apicryptosuite.CryptoSuite) error {
	if atomic.LoadInt32(&initialized) > 0 {
		return errors.New("default crypto suite is already set")
	}
	return initSuite(newDefaultSuite)
}

//GetSHA256Opts returns options relating to SHA-256.
func GetSHA256Opts() apicryptosuite.HashOpts {
	return &bccsp.SHA256Opts{}
}

//GetSHAOpts returns options for computing SHA.
func GetSHAOpts() apicryptosuite.HashOpts {
	return &bccsp.SHAOpts{}
}

//GetECDSAP256KeyGenOpts returns options for ECDSA key generation with curve P-256.
func GetECDSAP256KeyGenOpts(ephemeral bool) apicryptosuite.KeyGenOpts {
	return &bccsp.ECDSAP256KeyGenOpts{Temporary: ephemeral}
}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkcs11

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	bccspPkcs11 "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/factory/pkcs11"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/pkcs11"
	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp/wrapper"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

//GetSuiteByConfig returns cryptosuite adaptor for bccsp loaded according to given config
func GetSuiteByConfig(config apiconfig.Config) (apicryptosuite.CryptoSuite, error) {
	// TODO: delete this check?
	if config.SecurityProvider() != "PKCS11" {
		return nil, errors.Errorf("Unsupported BCCSP Provider: %s", config.SecurityProvider())
	}

	opts := getOptsByConfig(config)
	bccsp, err := getBCCSPFromOpts(opts)

	if err != nil {
		return nil, err
	}
	return &wrapper.CryptoSuite{BCCSP: bccsp}, nil
}

func getBCCSPFromOpts(config *pkcs11.PKCS11Opts) (bccsp.BCCSP, error) {
	f := &bccspPkcs11.PKCS11Factory{}

	csp, err := f.Get(config)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not initialize BCCSP %s", f.Name())
	}
	return csp, nil
}

//getOptsByConfig Returns Factory opts for given SDK config
func getOptsByConfig(c apiconfig.Config) *pkcs11.PKCS11Opts {
	pkks := pkcs11.FileKeystoreOpts{KeyStorePath: c.KeyStorePath()}
	opts := &pkcs11.PKCS11Opts{
		SecLevel:     c.SecurityLevel(),
		HashFamily:   c.SecurityAlgorithm(),
		Ephemeral:    c.Ephemeral(),
		FileKeystore: &pkks,
		Library:      c.SecurityProviderLibPath(),
		Pin:          c.SecurityProviderPin(),
		Label:        c.SecurityProviderLabel(),
		SoftVerify:   c.SoftVerify(),
	}
	logger.Debug("Initialized PKCS11 cryptosuite")

	return opts
}

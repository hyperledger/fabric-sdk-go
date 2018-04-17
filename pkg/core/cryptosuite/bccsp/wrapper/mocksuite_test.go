/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	bccspSw "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/factory/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
)

//getSuiteByConfig returns cryptosuite adaptor for bccsp loaded according to given config
func getSuiteByConfig(config core.CryptoSuiteConfig) (core.CryptoSuite, error) {
	opts := getOptsByConfig(config)
	bccsp, err := getBCCSPFromOpts(opts)

	if err != nil {
		return nil, err
	}
	return &CryptoSuite{BCCSP: bccsp}, nil
}

func getBCCSPFromOpts(config *bccspSw.SwOpts) (bccsp.BCCSP, error) {
	f := &bccspSw.SWFactory{}

	return f.Get(config)
}

//getOptsByConfig Returns Factory opts for given SDK config
func getOptsByConfig(c core.CryptoSuiteConfig) *bccspSw.SwOpts {
	// TODO: delete this check
	if c.SecurityProvider() != "SW" {
		panic(fmt.Sprintf("Unsupported BCCSP Provider: %s", c.SecurityProvider()))
	}

	opts := &bccspSw.SwOpts{
		HashFamily: c.SecurityAlgorithm(),
		SecLevel:   c.SecurityLevel(),
		FileKeystore: &bccspSw.FileKeystoreOpts{
			KeyStorePath: c.KeyStorePath(),
		},
	}

	return opts
}

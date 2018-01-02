// +build !pkcs11

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bccsp

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	bccspFactory "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/factory"
)

//GetOptsByConfig Returns Factory opts for given SDK config
func GetOptsByConfig(c apiconfig.Config) *bccspFactory.FactoryOpts {
	var opts *bccspFactory.FactoryOpts

	switch c.SecurityProvider() {
	case "SW":
		opts = &bccspFactory.FactoryOpts{
			ProviderName: "SW",
			SwOpts: &bccspFactory.SwOpts{
				HashFamily: c.SecurityAlgorithm(),
				SecLevel:   c.SecurityLevel(),
				FileKeystore: &bccspFactory.FileKeystoreOpts{
					KeyStorePath: c.KeyStorePath(),
				},
				Ephemeral: c.Ephemeral(),
			},
		}
		logger.Debug("Initialized SW ")
		bccspFactory.InitFactories(opts)
		return opts

	default:
		panic(fmt.Sprintf("Unsupported BCCSP Provider: %s", c.SecurityProvider()))
	}
}

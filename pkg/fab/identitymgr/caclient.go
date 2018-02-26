/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identitymgr

import (
	"github.com/pkg/errors"

	calib "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/lib"
	config "github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/urlutil"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

// initCAClient initializes a Fabric CA client instance.
// Initialization is lazy, so the client is not required to connect to CA
// in order to transact with Fabric.
func (im *IdentityManager) initCAClient() error {
	if im.caClient == nil {
		caClient, err := newCAClient(im.orgName, im.config, im.cryptoSuite)
		if err != nil {
			return errors.Wrapf(err, "failed to initialie Fabric CA client")
		}
		im.caClient = caClient
		caConfig, err := im.config.CAConfig(im.orgName)
		if err != nil {
			return errors.Wrapf(err, "failed to get CA configurtion for msp: %s", im.orgName)
		}
		im.registrar = caConfig.Registrar
	}
	return nil
}

func newCAClient(org string, config config.Config, cryptoSuite core.CryptoSuite) (*calib.Client, error) {

	// Create new Fabric-ca client without configs
	c := &calib.Client{
		Config: &calib.ClientConfig{},
	}

	conf, err := config.CAConfig(org)
	if err != nil {
		return nil, err
	}

	if conf == nil {
		return nil, errors.Errorf("Orgnization %s have no corresponding CA in the configs", org)
	}

	//set server CAName
	c.Config.CAName = conf.CAName
	//set server URL
	c.Config.URL = urlutil.ToAddress(conf.URL)
	//certs file list
	c.Config.TLS.CertFiles, err = config.CAServerCertPaths(org)
	if err != nil {
		return nil, err
	}

	// set key file and cert file
	c.Config.TLS.Client.CertFile, err = config.CAClientCertPath(org)
	if err != nil {
		return nil, err
	}

	c.Config.TLS.Client.KeyFile, err = config.CAClientKeyPath(org)
	if err != nil {
		return nil, err
	}

	// get Client configs
	_, err = config.Client()
	if err != nil {
		return nil, err
	}

	//TLS flag enabled/disabled
	c.Config.TLS.Enabled = urlutil.IsTLSEnabled(conf.URL)
	c.Config.MSPDir = config.CAKeyStorePath()

	//Factory opts
	c.Config.CSP = cryptoSuite

	err = c.Init()
	if err != nil {
		return nil, errors.Wrap(err, "init failed")
	}

	return c, nil
}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package membership

import (
	"crypto/x509"
	"encoding/pem"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	mb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

type identityImpl struct {
	mspManager msp.MSPManager
}

// Context holds the providers
type Context struct {
	core.Providers
}

// New member identity
func New(ctx Context, cfg fab.ChannelCfg) (fab.ChannelMembership, error) {
	m, err := createMSPManager(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &identityImpl{mspManager: m}, nil
}

func (i *identityImpl) Validate(serializedID []byte) error {
	id, err := i.mspManager.DeserializeIdentity(serializedID)
	if err != nil {
		return err
	}

	return id.Validate()
}

func (i *identityImpl) Verify(serializedID []byte, msg []byte, sig []byte) error {
	id, err := i.mspManager.DeserializeIdentity(serializedID)
	if err != nil {
		return err
	}

	return id.Verify(msg, sig)
}

func createMSPManager(ctx Context, cfg fab.ChannelCfg) (msp.MSPManager, error) {
	mspManager := msp.NewMSPManager()
	if len(cfg.MSPs()) > 0 {
		msps, err := loadMSPs(cfg.MSPs(), ctx.CryptoSuite())
		if err != nil {
			return nil, errors.WithMessage(err, "load MSPs from config failed")
		}

		if err := mspManager.Setup(msps); err != nil {
			return nil, errors.WithMessage(err, "MSPManager Setup failed")
		}

		for _, msp := range msps {
			for _, cert := range msp.GetTLSRootCerts() {
				addCertsToConfig(ctx.Config(), cert)
			}

			for _, cert := range msp.GetTLSIntermediateCerts() {
				addCertsToConfig(ctx.Config(), cert)
			}
		}
	}

	return mspManager, nil
}

func loadMSPs(mspConfigs []*mb.MSPConfig, cs core.CryptoSuite) ([]msp.MSP, error) {
	logger.Debugf("loadMSPs - start number of msps=%d", len(mspConfigs))

	msps := []msp.MSP{}
	for _, config := range mspConfigs {
		mspType := msp.ProviderType(config.Type)
		if mspType != msp.FABRIC {
			return nil, errors.Errorf("MSP type not supported: %v", mspType)
		}
		if len(config.Config) == 0 {
			return nil, errors.Errorf("MSP configuration missing the payload in the 'Config' property")
		}

		fabricConfig := &mb.FabricMSPConfig{}
		err := proto.Unmarshal(config.Config, fabricConfig)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshal FabricMSPConfig from config failed")
		}

		if fabricConfig.Name == "" {
			return nil, errors.New("MSP Configuration missing name")
		}

		// with this method we are only dealing with verifying MSPs, not local MSPs. Local MSPs are instantiated
		// from user enrollment materials (see User class). For verifying MSPs the root certificates are always
		// required
		if len(fabricConfig.RootCerts) == 0 {
			return nil, errors.New("MSP Configuration missing root certificates required for validating signing certificates")
		}

		// get the application org names
		var orgs []string
		orgUnits := fabricConfig.OrganizationalUnitIdentifiers
		for _, orgUnit := range orgUnits {
			logger.Debugf("loadMSPs - found org of :: %s", orgUnit.OrganizationalUnitIdentifier)
			orgs = append(orgs, orgUnit.OrganizationalUnitIdentifier)
		}

		// TODO: Do something with orgs
		// TODO: Configure MSP version (rather than MSP 1.0)
		newMSP, err := msp.NewBccspMsp(msp.MSPv1_0, cs)
		if err != nil {
			return nil, errors.Wrap(err, "instantiate MSP failed")
		}

		if err := newMSP.Setup(config); err != nil {
			return nil, errors.Wrap(err, "configure MSP failed")
		}

		mspID, _ := newMSP.GetIdentifier()
		logger.Debugf("loadMSPs - adding msp=%s", mspID)

		msps = append(msps, newMSP)
	}

	logger.Debugf("loadMSPs - loaded %d MSPs", len(msps))
	return msps, nil
}

//addCertsToConfig adds cert bytes to config TLSCACertPool
func addCertsToConfig(config core.Config, pemCerts []byte) {
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}
		config.TLSCACertPool(cert)
	}
}

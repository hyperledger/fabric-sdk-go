/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package membership

import (
	"crypto/x509"
	"encoding/pem"

	"strings"

	"github.com/golang/protobuf/proto"
	mb "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/verifier"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

type identityImpl struct {
	mspManager msp.MSPManager
	msps       []string
}

// Context holds the providers
type Context struct {
	core.Providers
	EndpointConfig fab.EndpointConfig
}

// New member identity
func New(ctx Context, cfg fab.ChannelCfg) (fab.ChannelMembership, error) {
	mspManager, mspNames, err := createMSPManager(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &identityImpl{mspManager: mspManager, msps: mspNames}, nil
}

func (i *identityImpl) Validate(serializedID []byte) error {
	err := areCertDatesValid(serializedID)
	if err != nil {
		logger.Errorf("Cert error %s", err)
		return err
	}

	id, err := i.mspManager.DeserializeIdentity(serializedID)
	if err != nil {
		logger.Errorf("failed to deserialize identity: %s", err)
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

func (i *identityImpl) ContainsMSP(msp string) bool {
	for _, v := range i.msps {
		if v == strings.ToLower(msp) {
			return true
		}
	}
	return false
}

func areCertDatesValid(serializedID []byte) error {

	sID := &mb.SerializedIdentity{}
	err := proto.Unmarshal(serializedID, sID)
	if err != nil {
		return errors.Wrap(err, "could not deserialize a SerializedIdentity")
	}

	bl, _ := pem.Decode(sID.IdBytes)
	if bl == nil {
		return errors.New("could not decode the PEM structure")
	}
	cert, err := x509.ParseCertificate(bl.Bytes)
	if err != nil {
		return err
	}
	err = verifier.ValidateCertificateDates(cert)
	if err != nil {
		logger.Warnf("Certificate error '%s' for cert '%v'", err, cert.SerialNumber)
		return err
	}
	return nil
}

func createMSPManager(ctx Context, cfg fab.ChannelCfg) (msp.MSPManager, []string, error) {
	mspManager := msp.NewMSPManager()
	var mspNames []string
	if len(cfg.MSPs()) > 0 {
		msps, err := loadMSPs(cfg.MSPs(), ctx.CryptoSuite())
		if err != nil {
			return nil, nil, errors.WithMessage(err, "load MSPs from config failed")
		}

		if err := mspManager.Setup(msps); err != nil {
			return nil, nil, errors.WithMessage(err, "MSPManager Setup failed")
		}

		certsByMsp := make(map[string][][]byte)
		for _, msp := range msps {
			mspName, err := msp.GetIdentifier()
			if err != nil {
				return nil, nil, errors.WithMessage(err, "MSPManager certpool setup failed")
			}
			certsByMsp[mspName] = append(msp.GetTLSRootCerts(), msp.GetTLSIntermediateCerts()...)
		}

		for mspName, certs := range certsByMsp {
			addCertsToConfig(ctx.EndpointConfig, certs)
			mspNames = append(mspNames, strings.ToLower(mspName))
		}
	}

	//To make sure tls cert pool is updated in advance with all the new certs being added,
	// to avoid delay in first endorsement connection with new peer
	_, err := ctx.EndpointConfig.TLSCACertPool().Get()
	if err != nil {
		return nil, nil, err
	}

	return mspManager, mspNames, nil
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

		fabricConfig, err := getFabricConfig(config)
		if err != nil {
			return nil, err
		}

		// get the application org names
		orgUnits := fabricConfig.OrganizationalUnitIdentifiers
		for _, orgUnit := range orgUnits {
			logger.Debugf("loadMSPs - found org of :: %s", orgUnit.OrganizationalUnitIdentifier)
		}

		// TODO: Do something with orgs
		// TODO: Configure MSP version
		mspOpts := msp.BCCSPNewOpts{
			NewBaseOpts: msp.NewBaseOpts{
				Version: msp.MSPv1_1,
			},
		}
		newMSP, err := msp.New(&mspOpts, cs)
		if err != nil {
			return nil, errors.Wrap(err, "instantiate MSP failed")
		}

		if err := newMSP.Setup(config); err != nil {
			return nil, errors.Wrap(err, "configure MSP failed")
		}

		mspID, err1 := newMSP.GetIdentifier()
		if err1 != nil {
			return nil, errors.Wrap(err1, "failed to get identifier")
		}
		logger.Debugf("loadMSPs - adding msp=%s", mspID)

		msps = append(msps, newMSP)
	}

	logger.Debugf("loadMSPs - loaded %d MSPs", len(msps))
	return msps, nil
}

func getFabricConfig(config *mb.MSPConfig) (*mb.FabricMSPConfig, error) {

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

	return fabricConfig, nil
}

//addCertsToConfig adds cert bytes to config TLSCACertPool
func addCertsToConfig(config fab.EndpointConfig, pemCertsList [][]byte) {

	if len(pemCertsList) == 0 {
		return
	}

	var certs []*x509.Certificate
	for _, pemCerts := range pemCertsList {
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
			err = verifier.ValidateCertificateDates(cert)
			if err != nil {
				logger.Warn("%v", err)
				continue
			}

			certs = append(certs, cert)
		}
	}

	config.TLSCACertPool().Add(certs...)
}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric-protos-go/msp"
	mspcfg "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
)

const (
	cacerts              = "cacerts"
	admincerts           = "admincerts"
	intermediatecerts    = "intermediatecerts"
	crlsfolder           = "crls"
	tlscacerts           = "tlscacerts"
	tlsintermediatecerts = "tlsintermediatecerts"
)

// GenerateMspDir generates a MSP directory, using values from the provided MSP config.
// The intended usage is within the scope of creating a genesis block. This means
// private keys are currently not handled.
func GenerateMspDir(mspDir string, config *msp.MSPConfig) error {

	if mspcfg.ProviderTypeToString(mspcfg.ProviderType(config.Type)) != "bccsp" {
		return fmt.Errorf("Unsupported MSP config type")
	}

	cfg := &msp.FabricMSPConfig{}
	err := proto.Unmarshal(config.Config, cfg)
	if err != nil {
		return err
	}

	type certDirDefinition struct {
		dir   string
		certs [][]byte
	}
	defs := []certDirDefinition{
		{cacerts, cfg.RootCerts},
		{admincerts, cfg.Admins},
		{intermediatecerts, cfg.IntermediateCerts},
		{tlscacerts, cfg.TlsRootCerts},
		{tlsintermediatecerts, cfg.TlsIntermediateCerts},
		{crlsfolder, cfg.RevocationList},
	}
	for _, d := range defs {
		errGen := generateCertDir(filepath.Join(mspDir, d.dir), d.certs)
		if errGen != nil {
			return errGen
		}
	}

	return err
}

func generateCertDir(certDir string, certs [][]byte) error {
	err := os.MkdirAll(certDir, 0750)
	if err != nil {
		return err
	}
	if len(certs) == 0 {
		return nil
	}
	for counter, certBytes := range certs {
		fileName := filepath.Join(certDir, "cert"+fmt.Sprintf("%d", counter)+".pem")
		err = ioutil.WriteFile(fileName, certBytes, 0640)
		if err != nil {
			return err
		}
	}
	return nil
}

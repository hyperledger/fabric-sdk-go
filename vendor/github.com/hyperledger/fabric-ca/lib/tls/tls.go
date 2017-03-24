/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

                 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/cloudflare/cfssl/log"
	"github.com/hyperledger/fabric-ca/util"
)

// ServerTLSConfig defines key material for a TLS server
type ServerTLSConfig struct {
	Enabled  bool   `help:"Enable TLS on the listening port"`
	CertFile string `def:"ca-cert.pem" help:"PEM-encoded TLS certificate file for server's listening port"`
	KeyFile  string `def:"ca-key.pem" help:"PEM-encoded TLS key for server's listening port"`
}

// ClientTLSConfig defines the key material for a TLS client
type ClientTLSConfig struct {
	Enabled       bool   `help:"Enable TLS for client connection"`
	CertFiles     string `help:"PEM-encoded comma separated list of trusted certificate files (e.g. root1.pem, root2.pem)"`
	CertFilesList []string
	Client        KeyCertFiles
}

// KeyCertFiles defines the files need for client on TLS
type KeyCertFiles struct {
	KeyFile  string `help:"PEM-encoded key file when mutual authentication is enabled"`
	CertFile string `help:"PEM-encoded certificate file when mutual authenticate is enabled"`
}

// GetClientTLSConfig creates a tls.Config object from certs and roots
func GetClientTLSConfig(cfg *ClientTLSConfig) (*tls.Config, error) {
	var certs []tls.Certificate

	log.Debugf("CA Files: %s\n", cfg.CertFiles)
	log.Debugf("Client Cert File: %s\n", cfg.Client.CertFile)
	log.Debugf("Client Key File: %s\n", cfg.Client.KeyFile)
	clientCert, err := tls.LoadX509KeyPair(cfg.Client.CertFile, cfg.Client.KeyFile)
	if err != nil {
		log.Debugf("Client Cert or Key not provided, if server requires mutual TLS, the connection will fail: %s", err)
	}

	certs = append(certs, clientCert)

	rootCAPool := x509.NewCertPool()

	if len(cfg.CertFilesList) == 0 {
		return nil, errors.New("No CA certificate files provided")
	}

	for _, cacert := range cfg.CertFilesList {
		caCert, err := ioutil.ReadFile(cacert)
		if err != nil {
			return nil, err
		}
		ok := rootCAPool.AppendCertsFromPEM(caCert)
		if !ok {
			return nil, fmt.Errorf("Failed to process certificate from file %s", cacert)
		}
	}

	config := &tls.Config{
		Certificates: certs,
		RootCAs:      rootCAPool,
	}

	return config, nil
}

// AbsTLSClient makes TLS client files absolute
func AbsTLSClient(cfg *ClientTLSConfig, configDir string) error {
	var err error

	for i := 0; i < len(cfg.CertFilesList); i++ {
		cfg.CertFilesList[i], err = util.MakeFileAbs(cfg.CertFilesList[i], configDir)
		if err != nil {
			return err
		}

	}

	cfg.Client.CertFile, err = util.MakeFileAbs(cfg.Client.CertFile, configDir)
	if err != nil {
		return err
	}

	cfg.Client.KeyFile, err = util.MakeFileAbs(cfg.Client.KeyFile, configDir)
	if err != nil {
		return err
	}

	return nil
}

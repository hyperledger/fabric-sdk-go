/*
Copyright IBM Corp All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package operations

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkinternal/pkg/comm"
)

type TLS struct {
	Enabled            bool
	CertFile           string
	KeyFile            string
	ClientCertRequired bool
	ClientCACertFiles  []string
}

func (t TLS) Config() (*tls.Config, error) {
	var tlsConfig *tls.Config

	if t.Enabled {
		cert, err := tls.LoadX509KeyPair(t.CertFile, t.KeyFile)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		for _, caPath := range t.ClientCACertFiles {
			caPem, err := ioutil.ReadFile(caPath)
			if err != nil {
				return nil, err
			}
			caCertPool.AppendCertsFromPEM(caPem)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			CipherSuites: comm.DefaultTLSCipherSuites,
			ClientCAs:    caCertPool,
		}
		if t.ClientCertRequired {
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven
		}
	}

	return tlsConfig, nil
}

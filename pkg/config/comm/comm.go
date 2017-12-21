/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"crypto/tls"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	cutil "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
)

// TLSConfig returns the appropriate config for TLS including the root CAs,
// certs for mutual TLS, and server host override
func TLSConfig(certificate string, serverhostoverride string, config apiconfig.Config) (*tls.Config, error) {
	certPool, _ := config.TLSCACertPool("")

	if len(certificate) == 0 && (certPool == nil || len(certPool.Subjects()) == 0) {
		return nil, errors.New("certificate is required")
	}

	tlsCaCertPool, err := config.TLSCACertPool(certificate)
	if err != nil {
		return nil, err
	}

	clientCerts, err := config.TLSClientCerts()
	if err != nil {
		return nil, errors.Errorf("Error loading cert/key pair for TLS client credentials: %v", err)
	}

	return &tls.Config{RootCAs: tlsCaCertPool, Certificates: clientCerts, ServerName: serverhostoverride}, nil
}

// TLSCertHash is a utility method to calculate the SHA256 hash of the configured certificate (for usage in channel headers)
func TLSCertHash(config apiconfig.Config) []byte {
	certs, err := config.TLSClientCerts()
	if err != nil || len(certs) == 0 {
		return nil
	}

	cert := certs[0]
	if len(cert.Certificate) == 0 {
		return nil
	}

	h := cutil.ComputeSHA256(cert.Certificate[0])
	return h
}

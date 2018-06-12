/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"crypto/tls"

	"crypto/x509"

	cutil "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// TLSConfig returns the appropriate config for TLS including the root CAs,
// certs for mutual TLS, and server host override. Works with certs loaded either from a path or embedded pem.
func TLSConfig(cert *x509.Certificate, serverName string, config fab.EndpointConfig) (*tls.Config, error) {
	certPool, err := config.TLSCACertPool().Get()
	if err != nil {
		return nil, err
	}

	if cert == nil && (certPool == nil || len(certPool.Subjects()) == 0) {
		//Return empty tls config if there is no cert provided or if certpool unavailable
		return &tls.Config{}, nil
	}

	tlsCaCertPool, err := config.TLSCACertPool().Get(cert)
	if err != nil {
		return nil, err
	}

	return &tls.Config{RootCAs: tlsCaCertPool, Certificates: config.TLSClientCerts(), ServerName: serverName}, nil
}

// TLSCertHash is a utility method to calculate the SHA256 hash of the configured certificate (for usage in channel headers)
func TLSCertHash(config fab.EndpointConfig) []byte {
	certs := config.TLSClientCerts()
	if len(certs) == 0 {
		return nil
	}

	cert := certs[0]
	if len(cert.Certificate) == 0 {
		return nil
	}

	h := cutil.ComputeSHA256(cert.Certificate[0])
	return h
}

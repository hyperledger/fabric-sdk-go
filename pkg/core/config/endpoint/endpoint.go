/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package endpoint

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"strings"

	"regexp"

	"github.com/pkg/errors"
)

// IsTLSEnabled is a generic function that expects a URL and verifies if it has
// a prefix HTTPS or GRPCS to return true for TLS Enabled URLs or false otherwise
func IsTLSEnabled(url string) bool {
	tlsURL := strings.ToLower(url)
	if strings.HasPrefix(tlsURL, "https://") || strings.HasPrefix(tlsURL, "grpcs://") {
		return true
	}
	return false
}

// ToAddress is a utility function to trim the GRPC protocol prefix as it is not needed by GO
// if the GRPC protocol is not found, the url is returned unchanged
func ToAddress(url string) string {
	if strings.HasPrefix(url, "grpc://") {
		return strings.TrimPrefix(url, "grpc://")
	}
	if strings.HasPrefix(url, "grpcs://") {
		return strings.TrimPrefix(url, "grpcs://")
	}
	return url
}

//AttemptSecured is a utility function which verifies URL and returns if secured connections needs to established
// for protocol 'grpcs' in URL returns true
// for protocol 'grpc' in URL returns false
// for no protocol mentioned, returns !allowInSecure
func AttemptSecured(url string, allowInSecure bool) bool {
	ok, err := regexp.MatchString(".*(?i)s://", url)
	if ok && err == nil {
		return true
	} else if strings.Contains(url, "://") {
		return false
	} else {
		return !allowInSecure
	}
}

// MutualTLSConfig Mutual TLS configurations
type MutualTLSConfig struct {
	Pem []string
	// Certfiles root certificates for TLS validation (Comma separated path list)
	Path string

	//Client TLS information
	Client TLSKeyPair
}

// TLSKeyPair contains the private key and certificate for TLS encryption
type TLSKeyPair struct {
	Key  TLSConfig
	Cert TLSConfig
}

// TLSConfig TLS configuration used in the sdk's configs.
type TLSConfig struct {
	// the following two fields are interchangeable.
	// If Path is available, then it will be used to load the cert
	// if Pem is available, then it has the raw data of the cert it will be used as-is
	// Certificate root certificate path
	// If both Path and Pem are available, pem takes the precedence
	Path string
	// Certificate actual content
	Pem string
	//bytes from Pem/Path
	bytes []byte
}

// Bytes returns the tls certificate as a byte array
func (cfg *TLSConfig) Bytes() []byte {
	return cfg.bytes
}

//LoadBytes preloads bytes from Pem/Path
//Pem takes precedence over Path
func (cfg *TLSConfig) LoadBytes() error {
	var err error
	if cfg.Pem != "" {
		cfg.bytes = []byte(cfg.Pem)
	} else if cfg.Path != "" {
		cfg.bytes, err = ioutil.ReadFile(cfg.Path)
		if err != nil {
			return errors.Wrapf(err, "failed to load pem bytes from path %s", cfg.Path)
		}
	}
	return nil
}

// TLSCert returns the tls certificate as a *x509.Certificate by loading it either from the embedded Pem or Path
func (cfg *TLSConfig) TLSCert() (*x509.Certificate, bool, error) {

	block, _ := pem.Decode(cfg.bytes)

	if block != nil {
		pub, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, false, errors.Wrap(err, "certificate parsing failed")
		}

		return pub, true, nil
	}

	//no cert found and there is no error
	return nil, false, nil
}

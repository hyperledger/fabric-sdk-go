/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cryptoutil

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/pkg/errors"

	factory "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/sdkpatch/cryptosuitebridge"
)

var logger = logging.NewLogger("fabsdk/core")

// GetPrivateKeyFromCert will return private key represented by SKI in cert's public key
func GetPrivateKeyFromCert(cert []byte, cs core.CryptoSuite) (core.Key, error) {

	// get the public key in the right format
	certPubK, err := GetPublicKeyFromCert(cert, cs)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to import certificate's public key")
	}

	if certPubK == nil || certPubK.SKI() == nil {
		return nil, errors.New("Failed to get SKI")
	}

	// Get the key given the SKI value
	key, err := cs.GetKey(certPubK.SKI())
	if err != nil {
		return nil, errors.WithMessage(err, "Could not find matching key for SKI")
	}

	if key != nil && !key.Private() {
		return nil, errors.Errorf("Found key is not private, SKI: %s", certPubK.SKI())
	}

	return key, nil
}

// GetPublicKeyFromCert will return public key the from cert
func GetPublicKeyFromCert(cert []byte, cs core.CryptoSuite) (core.Key, error) {

	dcert, _ := pem.Decode(cert)
	if dcert == nil {
		return nil, errors.Errorf("Unable to decode cert bytes [%v]", cert)
	}

	x509Cert, err := x509.ParseCertificate(dcert.Bytes)
	if err != nil {
		return nil, errors.Errorf("Unable to parse cert from decoded bytes: %s", err)
	}

	// get the public key in the right format
	key, err := cs.KeyImport(x509Cert, factory.GetX509PublicKeyImportOpts(true))
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to import certificate's public key")
	}

	return key, nil
}

// X509KeyPair will return cer/key pair used for mutual TLS
func X509KeyPair(certPEMBlock []byte, pk core.Key, cs core.CryptoSuite) (tls.Certificate, error) {

	fail := func(err error) (tls.Certificate, error) { return tls.Certificate{}, err }

	var cert tls.Certificate
	for {
		var certDERBlock *pem.Block
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		} else {
			logger.Debugf("Skipping block type: %s", certDERBlock.Type)
		}
	}

	if len(cert.Certificate) == 0 {
		return fail(errors.New("No certs available from bytes"))
	}

	// We are parsing public key for TLS to find its type
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fail(err)
	}

	switch x509Cert.PublicKey.(type) {
	case *ecdsa.PublicKey:
		cert.PrivateKey = &PrivateKey{cs, pk, &ecdsa.PublicKey{}}
	default:
		return fail(errors.New("tls: unknown public key algorithm"))
	}

	return cert, nil
}

//PrivateKey is signer implementation for golang client TLS
type PrivateKey struct {
	cryptoSuite core.CryptoSuite
	key         core.Key
	publicKey   crypto.PublicKey
}

// Public returns the public key corresponding to private key
func (priv *PrivateKey) Public() crypto.PublicKey {
	return priv.publicKey
}

// Sign signs msg with priv, reading randomness from rand. If opts is a
// *PSSOptions then the PSS algorithm will be used, otherwise PKCS#1 v1.5 will
// be used. This method is intended to support keys where the private part is
// kept in, for example, a hardware module.
func (priv *PrivateKey) Sign(rand io.Reader, msg []byte, opts crypto.SignerOpts) ([]byte, error) {
	if priv.cryptoSuite == nil {
		return nil, errors.New("Crypto suite not set")
	}

	if priv.key == nil {
		return nil, errors.New("Private key not set")
	}

	return priv.cryptoSuite.Sign(priv.key, msg, opts)
}

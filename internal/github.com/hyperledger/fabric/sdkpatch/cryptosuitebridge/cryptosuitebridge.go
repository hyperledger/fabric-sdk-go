/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package cryptosuitebridge

import (
	"crypto"
	"crypto/ecdsa"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	cspsigner "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/signer"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/utils"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
)

const (
	ECDSA            = bccsp.ECDSA
	ECDSAP256        = bccsp.ECDSAP256
	ECDSAP384        = bccsp.ECDSAP384
	ECDSAReRand      = bccsp.ECDSAReRand
	AES              = bccsp.AES
	AES128           = bccsp.AES128
	AES192           = bccsp.AES192
	AES256           = bccsp.AES256
	HMAC             = bccsp.HMAC
	HMACTruncated256 = bccsp.HMACTruncated256
	SHA              = bccsp.SHA
	SHA2             = bccsp.SHA2
	SHA3             = bccsp.SHA3
	SHA256           = bccsp.SHA256
	SHA384           = bccsp.SHA384
	SHA3_256         = bccsp.SHA3_256
	SHA3_384         = bccsp.SHA3_384
	X509Certificate  = bccsp.X509Certificate
)

// NewCspSigner is a bridge for bccsp signer.New call
func NewCspSigner(csp core.CryptoSuite, key core.Key) (crypto.Signer, error) {
	return cspsigner.New(csp, key)
}

//GetDefault creates new cryptosuite from bccsp factory default
func GetDefault() core.CryptoSuite {
	return cryptosuite.GetDefault()
}

//SignatureToLowS is a bridge for bccsp utils.SignatureToLowS()
func SignatureToLowS(k *ecdsa.PublicKey, signature []byte) ([]byte, error) {
	return utils.SignatureToLowS(k, signature)
}

//GetHashOpt is a bridge for bccsp util GetHashOpt
func GetHashOpt(hashFunction string) (core.HashOpts, error) {
	return bccsp.GetHashOpt(hashFunction)
}

//GetSHAOpts returns options for computing SHA.
func GetSHAOpts() core.HashOpts {
	return &bccsp.SHAOpts{}
}

//GetSHA256Opts returns options relating to SHA-256.
func GetSHA256Opts() core.HashOpts {
	return &bccsp.SHA256Opts{}
}

//GetSHA3256Opts returns options relating to SHA-256.
func GetSHA3256Opts() core.HashOpts {
	return &bccsp.SHA3_256Opts{}
}

// GetECDSAKeyGenOpts returns options for ECDSA key generation.
func GetECDSAKeyGenOpts(ephemeral bool) core.KeyGenOpts {
	return &bccsp.ECDSAKeyGenOpts{Temporary: ephemeral}
}

//GetECDSAP256KeyGenOpts returns options for ECDSA key generation with curve P-256.
func GetECDSAP256KeyGenOpts(ephemeral bool) core.KeyGenOpts {
	return &bccsp.ECDSAP256KeyGenOpts{Temporary: ephemeral}
}

//GetECDSAP384KeyGenOpts options for ECDSA key generation with curve P-384.
func GetECDSAP384KeyGenOpts(ephemeral bool) core.KeyGenOpts {
	return &bccsp.ECDSAP384KeyGenOpts{Temporary: ephemeral}
}

//GetX509PublicKeyImportOpts options for importing public keys from an x509 certificate
func GetX509PublicKeyImportOpts(ephemeral bool) core.KeyImportOpts {
	return &bccsp.X509PublicKeyImportOpts{Temporary: ephemeral}
}

//GetECDSAPrivateKeyImportOpts options for ECDSA secret key importation in DER format
// or PKCS#8 format.
func GetECDSAPrivateKeyImportOpts(ephemeral bool) core.KeyImportOpts {
	return &bccsp.ECDSAPrivateKeyImportOpts{Temporary: ephemeral}
}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package pkcs11wrapper

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"math/big"
)

type EcdsaKey struct {
	PubKey  *ecdsa.PublicKey
	PrivKey *ecdsa.PrivateKey
	SKI     SubjectKeyIdentifier
}

type SubjectKeyIdentifier struct {
	Sha1        string
	Sha1Bytes   []byte
	Sha256      string
	Sha256Bytes []byte
}

// SKI returns the subject key identifier of this key.
func (k *EcdsaKey) GenSKI() error {
	if k.PubKey == nil {
		return errors.New("PubKey is empty")
	}

	// Marshall the public key
	raw := elliptic.Marshal(k.PubKey.Curve, k.PubKey.X, k.PubKey.Y)

	// Hash it
	hash := sha256.New()
	_, err := hash.Write(raw)
	if err != nil {
		return errors.Wrap(err, "Failed to write hash")
	}
	k.SKI.Sha256Bytes = hash.Sum(nil)
	k.SKI.Sha256 = hex.EncodeToString(k.SKI.Sha256Bytes)

	hash = sha1.New()
	_, err = hash.Write(raw)
	if err != nil {
		return errors.Wrap(err, "Failed to write hash")
	}
	k.SKI.Sha1Bytes = hash.Sum(nil)
	k.SKI.Sha1 = hex.EncodeToString(k.SKI.Sha1Bytes)

	return nil
}

func (k *EcdsaKey) Generate(namedCurve string) (err error) {

	// generate private key
	switch namedCurve {
	case "P-224":
		k.PrivKey, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P-256":
		k.PrivKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P-384":
		k.PrivKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P-521":
		k.PrivKey, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		k.PrivKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	}

	// store public key
	k.PubKey = &k.PrivKey.PublicKey

	return
}

func (k *EcdsaKey) ImportPubKeyFromCertFile(file string) (err error) {

	certFile, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	certBlock, _ := pem.Decode(certFile)
	x509Cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return
	}

	k.PubKey = x509Cert.PublicKey.(*ecdsa.PublicKey)

	return
}

func (k *EcdsaKey) ImportPrivKeyFromFile(file string) (err error) {

	keyFile, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	keyBlock, _ := pem.Decode(keyFile)
	key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return
	}

	k.PrivKey = key.(*ecdsa.PrivateKey)
	k.PubKey = &k.PrivKey.PublicKey

	return
}

/* returns value for CKA_EC_PARAMS */
func GetECParamMarshaled(namedCurve string) (ecParamMarshaled []byte, err error) {

	// RFC 5480, 2.1.1.1. Named Curve
	//
	// secp224r1 OBJECT IDENTIFIER ::= {
	//   iso(1) identified-organization(3) certicom(132) curve(0) 33 }
	//
	// secp256r1 OBJECT IDENTIFIER ::= {
	//   iso(1) member-body(2) us(840) ansi-X9-62(10045) curves(3)
	//   prime(1) 7 }
	//
	// secp384r1 OBJECT IDENTIFIER ::= {
	//   iso(1) identified-organization(3) certicom(132) curve(0) 34 }
	//
	// secp521r1 OBJECT IDENTIFIER ::= {
	//   iso(1) identified-organization(3) certicom(132) curve(0) 35 }
	//
	// NB: secp256r1 is equivalent to prime256v1

	ecParamOID := asn1.ObjectIdentifier{}

	switch namedCurve {
	case "P-224":
		ecParamOID = asn1.ObjectIdentifier{1, 3, 132, 0, 33}
	case "P-256":
		ecParamOID = asn1.ObjectIdentifier{1, 2, 840, 10045, 3, 1, 7}
	case "P-384":
		ecParamOID = asn1.ObjectIdentifier{1, 3, 132, 0, 34}
	case "P-521":
		ecParamOID = asn1.ObjectIdentifier{1, 3, 132, 0, 35}
	}

	if len(ecParamOID) == 0 {
		err = fmt.Errorf("error with curve name: %s", namedCurve)
		return
	}

	ecParamMarshaled, err = asn1.Marshal(ecParamOID)
	return
}

func (k *EcdsaKey) SignMessage(message string) (signature string, err error) {

	// we should always hash the message before signing it
	// https://www.ietf.org/rfc/rfc4754.txt
	// https://tools.ietf.org/html/rfc5656#section-6.2.1
	//  +----------------+----------------+
	//  |   Curve Size   | Hash Algorithm |
	//	+----------------+----------------+
	//  |    b <= 256    |     SHA-256    |
	//  |                |                |
	//  | 256 < b <= 384 |     SHA-384    |
	//  |                |                |
	//  |     384 < b    |     SHA-512    |
	//	+----------------+----------------+
	bs := k.PrivKey.Params().BitSize
	var digest []byte

	switch {

	case bs <= 256:
		d := sha256.Sum256([]byte(message))
		digest = d[:]

	case bs > 256 && bs <= 384:
		d := sha512.Sum384([]byte(message))
		digest = d[:]

	case bs > 384:
		d := sha512.Sum512([]byte(message))
		digest = d[:]
	}

	// sign the hash
	// if the hash length is greater than the key length,
	// then only the first part of the hash that reaches the length of the key will be used
	r, s, err := ecdsa.Sign(rand.Reader, k.PrivKey, digest[:])
	if err != nil {
		return
	}

	signatureBytes := r.Bytes()
	signatureBytes = append(signatureBytes, s.Bytes()...)

	signature = hex.EncodeToString(signatureBytes)

	return
}

func (k *EcdsaKey) VerifySignature(message string, signature string) (verified bool) {

	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return
	}

	// we should always hash the message before signing it
	// https://www.ietf.org/rfc/rfc4754.txt
	bs := k.PrivKey.Params().BitSize
	var digest []byte

	switch {

	case bs <= 256:
		d := sha256.Sum256([]byte(message))
		digest = d[:]

	case bs > 256 && bs <= 384:
		d := sha512.Sum384([]byte(message))
		digest = d[:]

	case bs > 384:
		d := sha512.Sum512([]byte(message))
		digest = d[:]
	}

	// get curve byte size
	curveOrderByteSize := k.PubKey.Curve.Params().P.BitLen() / 8

	// extract r and s
	r, s := new(big.Int), new(big.Int)
	r.SetBytes(signatureBytes[:curveOrderByteSize])
	s.SetBytes(signatureBytes[curveOrderByteSize:])

	verified = ecdsa.Verify(k.PubKey, digest[:], r, s)

	return
}

func (k *EcdsaKey) DeriveSharedSecret(anotherPublicKey *ecdsa.PublicKey) (secret []byte, err error) {

	x, _ := k.PrivKey.Curve.ScalarMult(anotherPublicKey.X, anotherPublicKey.Y, k.PrivKey.D.Bytes())
	secret = x.Bytes()

	return
}

/*
Copyright IBM Corp. 2017 All Rights Reserved.

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

package util

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	_ "time" // for ocspSignerFromConfig

	_ "github.com/cloudflare/cfssl/cli" // for ocspSignerFromConfig
	"github.com/cloudflare/cfssl/config"
	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/helpers"
	"github.com/cloudflare/cfssl/log"
	_ "github.com/cloudflare/cfssl/ocsp" // for ocspSignerFromConfig
	"github.com/cloudflare/cfssl/signer"
	"github.com/cloudflare/cfssl/signer/local"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/factory"
	cspsigner "github.com/hyperledger/fabric/bccsp/signer"
	"github.com/hyperledger/fabric/bccsp/utils"
)

// GetDefaultBCCSP returns the default BCCSP
func GetDefaultBCCSP() bccsp.BCCSP {
	return factory.GetDefault()
}

// InitBCCSP initializes BCCSP
func InitBCCSP(optsPtr **factory.FactoryOpts, mspDir, homeDir string) (bccsp.BCCSP, error) {
	err := ConfigureBCCSP(optsPtr, mspDir, homeDir)
	if err != nil {
		return nil, err
	}
	csp, err := GetBCCSP(*optsPtr, homeDir)
	if err != nil {
		return nil, err
	}
	return csp, nil
}

// ConfigureBCCSP configures BCCSP, using
func ConfigureBCCSP(optsPtr **factory.FactoryOpts, mspDir, homeDir string) error {
	var err error
	if optsPtr == nil {
		return errors.New("nil argument not allowed")
	}
	opts := *optsPtr
	if opts == nil {
		opts = &factory.FactoryOpts{}
	}
	if opts.ProviderName == "" {
		opts.ProviderName = "SW"
	}
	if strings.ToUpper(opts.ProviderName) == "SW" {
		if opts.SwOpts == nil {
			opts.SwOpts = &factory.SwOpts{}
		}
		if opts.SwOpts.HashFamily == "" {
			opts.SwOpts.HashFamily = "SHA2"
		}
		if opts.SwOpts.SecLevel == 0 {
			opts.SwOpts.SecLevel = 256
		}
		if opts.SwOpts.FileKeystore == nil {
			opts.SwOpts.FileKeystore = &factory.FileKeystoreOpts{}
		}
		// The mspDir overrides the KeyStorePath; otherwise, if not set, set default
		if mspDir != "" {
			opts.SwOpts.FileKeystore.KeyStorePath = path.Join(mspDir, "keystore")
		} else if opts.SwOpts.FileKeystore.KeyStorePath == "" {
			opts.SwOpts.FileKeystore.KeyStorePath = path.Join("msp", "keystore")
		}
	}
	err = makeFileNamesAbsolute(opts, homeDir)
	if err != nil {
		return fmt.Errorf("Failed to make BCCSP files absolute: %s", err)
	}
	log.Debugf("Initializing BCCSP: %+v", opts)
	if opts.SwOpts != nil {
		log.Debugf("Initializing BCCSP with software options %+v", opts.SwOpts)
	}
	if opts.Pkcs11Opts != nil {
		log.Debugf("Initializing BCCSP with PKCS11 options %+v", opts.Pkcs11Opts)
	}
	// Init the BCCSP factories
	err = factory.InitFactories(opts)
	if err != nil {
		return fmt.Errorf("Failed to initialize BCCSP Factories: %s", err)
	}
	*optsPtr = opts
	return nil
}

// GetBCCSP returns BCCSP
func GetBCCSP(opts *factory.FactoryOpts, homeDir string) (bccsp.BCCSP, error) {

	// Get BCCSP from the opts
	csp, err := factory.GetBCCSPFromOpts(opts)
	if err != nil {
		return nil, fmt.Errorf("Failed to get BCCSP: %s [opts: %+v]", err, opts)
	}
	return csp, nil
}

// makeFileNamesAbsolute makes all relative file names associated with CSP absolute,
// relative to 'homeDir'.
func makeFileNamesAbsolute(opts *factory.FactoryOpts, homeDir string) error {
	var err error
	if opts != nil && opts.SwOpts != nil && opts.SwOpts.FileKeystore != nil {
		fks := opts.SwOpts.FileKeystore
		fks.KeyStorePath, err = MakeFileAbs(fks.KeyStorePath, homeDir)
	}
	return err
}

// BccspBackedSigner attempts to create a signer using csp bccsp.BCCSP. This csp could be SW (golang crypto)
// PKCS11 or whatever BCCSP-conformant library is configured
func BccspBackedSigner(caFile, keyFile string, policy *config.Signing, csp bccsp.BCCSP) (signer.Signer, error) {
	_, cspSigner, parsedCa, err := GetSignerFromCertFile(caFile, csp)
	if err != nil {
		// Fallback: attempt to read out of keyFile and import
		log.Debugf("No key found in BCCSP keystore, attempting fallback")
		var key bccsp.Key
		var signer crypto.Signer

		key, err = ImportBCCSPKeyFromPEM(keyFile, csp, false)
		if err != nil {
			return nil, fmt.Errorf("Could not find the private key in BCCSP keystore nor in keyfile %s: %s", keyFile, err)
		}

		signer, err = cspsigner.New(csp, key)
		if err != nil {
			return nil, fmt.Errorf("Failed initializing CryptoSigner: %s", err)
		}
		cspSigner = signer
	}

	signer, err := local.NewSigner(cspSigner, parsedCa, signer.DefaultSigAlgo(cspSigner), policy)
	if err != nil {
		return nil, fmt.Errorf("Failed to create new signer: %s", err.Error())
	}
	return signer, nil
}

// getBCCSPKeyOpts generates a key as specified in the request.
// This supports ECDSA and RSA.
func getBCCSPKeyOpts(kr csr.KeyRequest, ephemeral bool) (opts bccsp.KeyGenOpts, err error) {
	if kr == nil {
		return &bccsp.ECDSAKeyGenOpts{Temporary: ephemeral}, nil
	}
	log.Debugf("generate key from request: algo=%s, size=%d", kr.Algo(), kr.Size())
	switch kr.Algo() {
	case "rsa":
		switch kr.Size() {
		case 2048:
			return &bccsp.RSA2048KeyGenOpts{Temporary: ephemeral}, nil
		case 3072:
			return &bccsp.RSA3072KeyGenOpts{Temporary: ephemeral}, nil
		case 4096:
			return &bccsp.RSA4096KeyGenOpts{Temporary: ephemeral}, nil
		default:
			// Need to add a way to specify arbitrary RSA key size to bccsp
			return nil, fmt.Errorf("Invalid RSA key size: %d", kr.Size())
		}
	case "ecdsa":
		switch kr.Size() {
		case 256:
			return &bccsp.ECDSAP256KeyGenOpts{Temporary: ephemeral}, nil
		case 384:
			return &bccsp.ECDSAP384KeyGenOpts{Temporary: ephemeral}, nil
		case 521:
			// Need to add curve P521 to bccsp
			// return &bccsp.ECDSAP512KeyGenOpts{Temporary: false}, nil
			return nil, errors.New("Unsupported ECDSA key size: 521")
		default:
			return nil, fmt.Errorf("Invalid ECDSA key size: %d", kr.Size())
		}
	default:
		return nil, fmt.Errorf("Invalid algorithm: %s", kr.Algo())
	}
}

// GetSignerFromCert load private key represented by ski and return bccsp signer that conforms to crypto.Signer
func GetSignerFromCert(cert *x509.Certificate, csp bccsp.BCCSP) (bccsp.Key, crypto.Signer, error) {
	if csp == nil {
		return nil, nil, fmt.Errorf("CSP was not initialized")
	}
	// get the public key in the right format
	certPubK, err := csp.KeyImport(cert, &bccsp.X509PublicKeyImportOpts{Temporary: true})
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to import certificate's public key: %s", err.Error())
	}
	// Get the key given the SKI value
	privateKey, err := csp.GetKey(certPubK.SKI())
	if err != nil {
		return nil, nil, fmt.Errorf("Could not find matching private key for SKI: %s", err.Error())
	}
	// Construct and initialize the signer
	signer, err := cspsigner.New(csp, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to load ski from bccsp: %s", err.Error())
	}
	return privateKey, signer, nil
}

// GetSignerFromCertFile load skiFile and load private key represented by ski and return bccsp signer that conforms to crypto.Signer
func GetSignerFromCertFile(certFile string, csp bccsp.BCCSP) (bccsp.Key, crypto.Signer, *x509.Certificate, error) {
	// Load cert file
	certBytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Could not read certFile [%s]: %s", certFile, err.Error())
	}
	// Parse certificate
	parsedCa, err := helpers.ParseCertificatePEM(certBytes)
	if err != nil {
		return nil, nil, nil, err
	}
	// Get the signer from the cert
	key, cspSigner, err := GetSignerFromCert(parsedCa, csp)
	return key, cspSigner, parsedCa, err
}

// BCCSPKeyRequestGenerate generates keys through BCCSP
// somewhat mirroring to cfssl/req.KeyRequest.Generate()
func BCCSPKeyRequestGenerate(req *csr.CertificateRequest, myCSP bccsp.BCCSP) (bccsp.Key, crypto.Signer, error) {
	log.Infof("generating key: %+v", req.KeyRequest)
	keyOpts, err := getBCCSPKeyOpts(req.KeyRequest, false)
	if err != nil {
		return nil, nil, err
	}
	key, err := myCSP.KeyGen(keyOpts)
	if err != nil {
		return nil, nil, err
	}

	cspSigner, err := cspsigner.New(myCSP, key)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed initializing CryptoSigner: %s", err.Error())
	}
	return key, cspSigner, nil
}

// ImportBCCSPKeyFromPEM attempts to create a private BCCSP key from a pem file keyFile
func ImportBCCSPKeyFromPEM(keyFile string, myCSP bccsp.BCCSP, temporary bool) (bccsp.Key, error) {
	keyBuff, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	key, err := utils.PEMtoPrivateKey(keyBuff, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed parsing private key from %s: %s", keyFile, err.Error())
	}
	switch key.(type) {
	case *ecdsa.PrivateKey:
		priv, err := utils.PrivateKeyToDER(key.(*ecdsa.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("Failed to convert ECDSA private key from %s: %s", keyFile, err.Error())
		}
		sk, err := myCSP.KeyImport(priv, &bccsp.ECDSAPrivateKeyImportOpts{Temporary: temporary})
		if err != nil {
			return nil, fmt.Errorf("Failed to import ECDSA private key from %s: %s", keyFile, err.Error())
		}
		return sk, nil
	case *rsa.PrivateKey:
		return nil, fmt.Errorf("Failed to import RSA key from %s; RSA private key import is not supported", keyFile)
	default:
		return nil, fmt.Errorf("Failed to import key from %s: invalid secret key type", keyFile)
	}
}

// LoadX509KeyPair reads and parses a public/private key pair from a pair
// of files. The files must contain PEM encoded data. The certificate file
// may contain intermediate certificates following the leaf certificate to
// form a certificate chain. On successful return, Certificate.Leaf will
// be nil because the parsed form of the certificate is not retained.
//
// This function originated from crypto/tls/tls.go and was adapted to use a
// BCCSP Signer
func LoadX509KeyPair(certFile, keyFile string, csp bccsp.BCCSP) (*tls.Certificate, error) {

	certPEMBlock, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	cert := &tls.Certificate{}
	var skippedBlockTypes []string
	for {
		var certDERBlock *pem.Block
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		} else {
			skippedBlockTypes = append(skippedBlockTypes, certDERBlock.Type)
		}
	}

	if len(cert.Certificate) == 0 {
		if len(skippedBlockTypes) == 0 {
			return nil, fmt.Errorf("Failed to find PEM block in file %s", certFile)
		}
		if len(skippedBlockTypes) == 1 && strings.HasSuffix(skippedBlockTypes[0], "PRIVATE KEY") {
			return nil, fmt.Errorf("Failed to find certificate PEM data in file %s, but did find a private key; PEM inputs may have been switched", certFile)
		}
		return nil, fmt.Errorf("Failed to find \"CERTIFICATE\" PEM block in file %s after skipping PEM blocks of the following types: %v", certFile, skippedBlockTypes)
	}

	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, err
	}

	_, cert.PrivateKey, err = GetSignerFromCert(x509Cert, csp)
	if err != nil {
		if keyFile != "" {
			log.Debugf("Could not load TLS certificate with BCCSP: %s", err)
			log.Debugf("Attempting fallback with certfile %s and keyfile %s", certFile, keyFile)
			fallbackCerts, err := tls.LoadX509KeyPair(certFile, keyFile)
			if err != nil {
				return nil, fmt.Errorf("Could not get the private key %s that matches %s: %s", keyFile, certFile, err)
			}
			cert = &fallbackCerts
		} else {
			return nil, fmt.Errorf("Could not load TLS certificate with BCCSP: %s", err)
		}

	}

	return cert, nil
}

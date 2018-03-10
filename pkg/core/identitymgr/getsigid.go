/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identitymgr

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/cryptoutil"

	fabricCaUtil "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
	"github.com/pkg/errors"
)

func newUser(userData msp.UserData, cryptoSuite core.CryptoSuite) (*User, error) {
	pubKey, err := cryptoutil.GetPublicKeyFromCert(userData.EnrollmentCertificate, cryptoSuite)
	if err != nil {
		return nil, errors.WithMessage(err, "fetching public key from cert failed")
	}
	pk, err := cryptoSuite.GetKey(pubKey.SKI())
	if err != nil {
		return nil, errors.WithMessage(err, "cryptoSuite GetKey failed")
	}
	u := &User{
		mspID: userData.MspID,
		name:  userData.Name,
		enrollmentCertificate: userData.EnrollmentCertificate,
		privateKey:            pk,
	}
	return u, nil
}

// NewUser creates a User instance
func (mgr *IdentityManager) NewUser(userData msp.UserData) (*User, error) {
	return newUser(userData, mgr.cryptoSuite)
}

func (mgr *IdentityManager) loadUserFromStore(userName string) (msp.User, error) {
	if mgr.userStore == nil {
		return nil, msp.ErrUserNotFound
	}
	var user msp.User
	userData, err := mgr.userStore.Load(msp.UserIdentifier{MspID: mgr.orgMspID, Name: userName})
	if err != nil {
		return nil, err
	}
	user, err = mgr.NewUser(userData)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetSigningIdentity returns a signing identity for the given user name
func (mgr *IdentityManager) GetSigningIdentity(userName string) (*msp.SigningIdentity, error) {
	user, err := mgr.GetUser(userName)
	if err != nil {
		return nil, err
	}
	signingIdentity := &msp.SigningIdentity{MspID: user.MspID(), PrivateKey: user.PrivateKey(), EnrollmentCert: user.EnrollmentCertificate()}
	return signingIdentity, nil
}

// GetUser returns a user for the given user name
func (mgr *IdentityManager) GetUser(userName string) (msp.User, error) {

	u, err := mgr.loadUserFromStore(userName)
	if err != nil {
		if err != msp.ErrUserNotFound {
			return nil, errors.WithMessage(err, "getting private key from cert failed")
		}
		// Not found, continue
	}

	if u == nil {
		certBytes, err := mgr.getEmbeddedCertBytes(userName)
		if err != nil && err != msp.ErrUserNotFound {
			return nil, errors.WithMessage(err, "fetching embedded cert failed")
		}
		if certBytes == nil {
			certBytes, err = mgr.getCertBytesFromCertStore(userName)
			if err != nil && err != msp.ErrUserNotFound {
				return nil, errors.WithMessage(err, "fetching cert from store failed")
			}
		}
		if certBytes == nil {
			return nil, msp.ErrUserNotFound
		}
		privateKey, err := mgr.getEmbeddedPrivateKey(userName)
		if err != nil {
			return nil, errors.WithMessage(err, "fetching embedded private key failed")
		}
		if privateKey == nil {
			privateKey, err = mgr.getPrivateKeyFromCert(userName, certBytes)
			if err != nil {
				return nil, errors.WithMessage(err, "getting private key from cert failed")
			}
		}
		if privateKey == nil {
			return nil, fmt.Errorf("unable to find private key for user [%s]", userName)
		}
		mspID, err := mgr.config.MspID(mgr.orgName)
		if err != nil {
			return nil, errors.WithMessage(err, "MSP ID config read failed")
		}
		u = &User{
			mspID: mspID,
			name:  userName,
			enrollmentCertificate: certBytes,
			privateKey:            privateKey,
		}
	}
	return u, nil
}

func (mgr *IdentityManager) getEmbeddedCertBytes(userName string) ([]byte, error) {
	certPem := mgr.embeddedUsers[strings.ToLower(userName)].Cert.Pem
	certPath := mgr.embeddedUsers[strings.ToLower(userName)].Cert.Path

	if certPem == "" && certPath == "" {
		return nil, msp.ErrUserNotFound
	}

	var pemBytes []byte
	var err error

	if certPem != "" {
		pemBytes = []byte(certPem)
	} else if certPath != "" {
		pemBytes, err = ioutil.ReadFile(certPath)
		if err != nil {
			return nil, errors.WithMessage(err, "reading cert from embedded path failed")
		}
	}

	return pemBytes, nil
}

func (mgr *IdentityManager) getEmbeddedPrivateKey(userName string) (core.Key, error) {
	keyPem := mgr.embeddedUsers[strings.ToLower(userName)].Key.Pem
	keyPath := mgr.embeddedUsers[strings.ToLower(userName)].Key.Path

	var privateKey core.Key
	var pemBytes []byte
	var err error

	if keyPem != "" {
		// Try importing from the Embedded Pem
		pemBytes = []byte(keyPem)
	} else if keyPath != "" {
		// Try importing from the Embedded Path
		_, err := os.Stat(keyPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, errors.WithMessage(err, "OS stat embedded path failed")
			}
			// file doesn't exist, continue
		} else {
			// file exists, try to read it
			pemBytes, err = ioutil.ReadFile(keyPath)
			if err != nil {
				return nil, errors.WithMessage(err, "reading private key from embedded path failed")
			}
		}
	}

	if pemBytes != nil {
		// Try the crypto provider as a SKI
		privateKey, err = mgr.cryptoSuite.GetKey(pemBytes)
		if err != nil || privateKey == nil {
			// Try as a pem
			privateKey, err = fabricCaUtil.ImportBCCSPKeyFromPEMBytes(pemBytes, mgr.cryptoSuite, true)
			if err != nil {
				return nil, errors.Wrapf(err, "import private key failed %v", keyPem)
			}
		}
	}

	return privateKey, nil
}

func (mgr *IdentityManager) getPrivateKeyPemFromKeyStore(userName string, ski []byte) ([]byte, error) {
	if mgr.mspPrivKeyStore == nil {
		return nil, nil
	}
	key, err := mgr.mspPrivKeyStore.Load(
		&msp.PrivKeyKey{
			MspID:    mgr.orgMspID,
			UserName: userName,
			SKI:      ski,
		})
	if err != nil {
		return nil, err
	}
	keyBytes, ok := key.([]byte)
	if !ok {
		return nil, errors.New("key from store is not []byte")
	}
	return keyBytes, nil
}

func (mgr *IdentityManager) getCertBytesFromCertStore(userName string) ([]byte, error) {
	if mgr.mspCertStore == nil {
		return nil, msp.ErrUserNotFound
	}
	cert, err := mgr.mspCertStore.Load(&msp.CertKey{
		MspID:    mgr.orgMspID,
		UserName: userName,
	})
	if err != nil {
		if err == core.ErrKeyValueNotFound {
			return nil, msp.ErrUserNotFound
		}
		return nil, err
	}
	certBytes, ok := cert.([]byte)
	if !ok {
		return nil, errors.New("cert from store is not []byte")
	}
	return certBytes, nil
}

func (mgr *IdentityManager) getPrivateKeyFromCert(userName string, cert []byte) (core.Key, error) {
	if cert == nil {
		return nil, errors.New("cert is nil")
	}
	pubKey, err := cryptoutil.GetPublicKeyFromCert(cert, mgr.cryptoSuite)
	if err != nil {
		return nil, errors.WithMessage(err, "fetching public key from cert failed")
	}
	privKey, err := mgr.getPrivateKeyFromKeyStore(userName, pubKey.SKI())
	if err == nil {
		return privKey, nil
	}
	if err != core.ErrKeyValueNotFound {
		return nil, errors.WithMessage(err, "fetching private key from key store failed")
	}
	return mgr.cryptoSuite.GetKey(pubKey.SKI())
}

func (mgr *IdentityManager) getPrivateKeyFromKeyStore(userName string, ski []byte) (core.Key, error) {
	pemBytes, err := mgr.getPrivateKeyPemFromKeyStore(userName, ski)
	if err != nil {
		return nil, err
	}
	if pemBytes != nil {
		return fabricCaUtil.ImportBCCSPKeyFromPEMBytes(pemBytes, mgr.cryptoSuite, true)
	}
	return nil, core.ErrKeyValueNotFound
}

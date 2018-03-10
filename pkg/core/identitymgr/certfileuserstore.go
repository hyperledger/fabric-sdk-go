/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identitymgr

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/keyvaluestore"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
	"github.com/pkg/errors"
)

// CertFileUserStore stores each user in a separate file.
// Only user's enrollment cert is stored, in pem format.
// File naming is <user>@<org>-cert.pem
type CertFileUserStore struct {
	store core.KVStore
}

func userIdentifierFromUser(user msp.UserData) msp.UserIdentifier {
	return msp.UserIdentifier{
		MspID: user.MspID,
		Name:  user.Name,
	}
}

func storeKeyFromUserIdentifier(key msp.UserIdentifier) string {
	return key.Name + "@" + key.MspID + "-cert.pem"
}

// NewCertFileUserStore1 creates a new instance of CertFileUserStore
func NewCertFileUserStore1(store core.KVStore) (*CertFileUserStore, error) {
	return &CertFileUserStore{
		store: store,
	}, nil
}

// NewCertFileUserStore creates a new instance of CertFileUserStore
func NewCertFileUserStore(path string) (*CertFileUserStore, error) {
	if path == "" {
		return nil, errors.New("path is empty")
	}
	store, err := keyvaluestore.New(&keyvaluestore.FileKeyValueStoreOptions{
		Path: path,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "user store creation failed")
	}
	return NewCertFileUserStore1(store)
}

// Load returns the User stored in the store for a key.
func (s *CertFileUserStore) Load(key msp.UserIdentifier) (msp.UserData, error) {
	var userData msp.UserData
	cert, err := s.store.Load(storeKeyFromUserIdentifier(key))
	if err != nil {
		if err == core.ErrKeyValueNotFound {
			return userData, msp.ErrUserNotFound
		}
		return userData, err
	}
	certBytes, ok := cert.([]byte)
	if !ok {
		return userData, errors.New("user is not of proper type")
	}
	userData = msp.UserData{
		MspID: key.MspID,
		Name:  key.Name,
		EnrollmentCertificate: certBytes,
	}
	return userData, nil
}

// Store stores a User into store
func (s *CertFileUserStore) Store(user msp.UserData) error {
	key := storeKeyFromUserIdentifier(msp.UserIdentifier{MspID: user.MspID, Name: user.Name})
	return s.store.Store(key, user.EnrollmentCertificate)
}

// Delete deletes a User from store
func (s *CertFileUserStore) Delete(key msp.UserIdentifier) error {
	return s.store.Delete(storeKeyFromUserIdentifier(key))
}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/keyvaluestore"
	"github.com/pkg/errors"
)

// CertFileUserStore stores each user in a separate file.
// Only user's enrollment cert is stored, in pem format.
// File naming is <user>@<org>-cert.pem
type CertFileUserStore struct {
	store core.KVStore
}

func storeKeyFromUserIdentifier(key msp.IdentityIdentifier) string {
	return key.ID + "@" + key.MSPID + "-cert.pem"
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
func (s *CertFileUserStore) Load(key msp.IdentityIdentifier) (*msp.UserData, error) {
	cert, err := s.store.Load(storeKeyFromUserIdentifier(key))
	if err != nil {
		if err == core.ErrKeyValueNotFound {
			return nil, msp.ErrUserNotFound
		}
		return nil, err
	}
	certBytes, ok := cert.([]byte)
	if !ok {
		return nil, errors.New("user is not of proper type")
	}
	userData := &msp.UserData{
		MSPID:                 key.MSPID,
		ID:                    key.ID,
		EnrollmentCertificate: certBytes,
	}
	return userData, nil
}

// Store stores a User into store
func (s *CertFileUserStore) Store(user *msp.UserData) error {
	key := storeKeyFromUserIdentifier(msp.IdentityIdentifier{MSPID: user.MSPID, ID: user.ID})
	return s.store.Store(key, user.EnrollmentCertificate)
}

// Delete deletes a User from store
func (s *CertFileUserStore) Delete(key msp.IdentityIdentifier) error {
	return s.store.Delete(storeKeyFromUserIdentifier(key))
}

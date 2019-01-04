/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
)

// MemoryUserStore is in-memory implementation of UserStore
type MemoryUserStore struct {
	store map[string][]byte
}

// NewMemoryUserStore creates a new MemoryUserStore instance
func NewMemoryUserStore() *MemoryUserStore {
	store := make(map[string][]byte)
	return &MemoryUserStore{store: store}
}

// Store stores a user into store
func (s *MemoryUserStore) Store(user *msp.UserData) error {
	s.store[user.ID+"@"+user.MSPID] = user.EnrollmentCertificate
	return nil
}

// Load loads a user from store
func (s *MemoryUserStore) Load(id msp.IdentityIdentifier) (*msp.UserData, error) {
	cert, ok := s.store[id.ID+"@"+id.MSPID]
	if !ok {
		return nil, msp.ErrUserNotFound
	}
	userData := msp.UserData{
		ID:                    id.ID,
		MSPID:                 id.MSPID,
		EnrollmentCertificate: cert,
	}
	return &userData, nil
}

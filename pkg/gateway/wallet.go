/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"encoding/json"

	"github.com/golang/protobuf/proto"
	pb_msp "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/pkg/errors"
)

type wallet interface {
	Put(label string, id Identity) error
	Get(label string) (Identity, error)
	Remove(label string) error
	Exists(label string) bool
	List() ([]string, error)
}

// A Wallet stores identity information used to connect to a Hyperledger Fabric network.
// Instances are created using factory methods on the implementing objects.
type Wallet struct {
	store WalletStore
}

// Put an identity into the wallet
//  Parameters:
//  label specifies the name to be associated with the identity.
//  id specifies the identity to store in the wallet.
//
func (w *Wallet) Put(label string, id Identity) error {
	content, err := id.toJSON()
	if err != nil {
		return err
	}

	return w.store.Put(label, content)
}

// Get an identity from the wallet. The implementation class of the identity object will vary depending on its type.
//  Parameters:
//  label specifies the name of the identity in the wallet.
//
//  Returns:
//  The identity object.
func (w *Wallet) Get(label string) (Identity, error) {
	content, err := w.store.Get(label)

	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, errors.Wrap(err, "Invalid identity format")
	}

	idType, ok := data["type"].(string)

	if !ok {
		return nil, errors.New("Invalid identity format: missing type property")
	}

	var id Identity

	switch idType {
	case x509Type:
		id = &X509Identity{}
	default:
		return nil, errors.New("Invalid identity format: unsupported identity type: " + idType)
	}

	return id.fromJSON(content)
}

// List returns the labels of all identities in the wallet.
//
//  Returns:
//  A list of identity labels in the wallet.
func (w *Wallet) List() ([]string, error) {
	return w.store.List()
}

// Exists tests whether the wallet contains an identity for the given label.
//  Parameters:
//  label specifies the name of the identity in the wallet.
//
//  Returns:
//  True if the named identity is in the wallet.
func (w *Wallet) Exists(label string) bool {
	return w.store.Exists(label)
}

// Remove an identity from the wallet. If the identity does not exist, this method does nothing.
//  Parameters:
//  label specifies the name of the identity in the wallet.
func (w *Wallet) Remove(label string) error {
	return w.store.Remove(label)
}

type walletmsp struct {
}

func (f *walletmsp) CreateUserStore(config msp.IdentityConfig) (msp.UserStore, error) {
	return nil, nil
}

func (f *walletmsp) CreateIdentityManagerProvider(config fab.EndpointConfig, cryptoProvider core.CryptoSuite, userStore msp.UserStore) (msp.IdentityManagerProvider, error) {
	return nil, nil
}

// walletIdentity is a representation of a Fabric User
type walletIdentity struct {
	id                    string
	mspID                 string
	enrollmentCertificate []byte
	privateKey            core.Key
}

// Identifier returns walletIdentity identifier
func (u *walletIdentity) Identifier() *msp.IdentityIdentifier {
	return &msp.IdentityIdentifier{MSPID: u.mspID, ID: u.id}
}

// Verify a signature over some message using this identity as reference
func (u *walletIdentity) Verify(msg []byte, sig []byte) error {
	return errors.New("not implemented")
}

// Serialize converts an identity to bytes
func (u *walletIdentity) Serialize() ([]byte, error) {
	serializedIdentity := &pb_msp.SerializedIdentity{
		Mspid:   u.mspID,
		IdBytes: u.enrollmentCertificate,
	}
	identity, err := proto.Marshal(serializedIdentity)
	if err != nil {
		return nil, errors.Wrap(err, "marshal serializedIdentity failed")
	}
	return identity, nil
}

// EnrollmentCertificate Returns the underlying ECert representing this walletIdentity’s identity.
func (u *walletIdentity) EnrollmentCertificate() []byte {
	return u.enrollmentCertificate
}

// PrivateKey returns the crypto suite representation of the private key
func (u *walletIdentity) PrivateKey() core.Key {
	return u.privateKey
}

// PublicVersion returns the public parts of this identity
func (u *walletIdentity) PublicVersion() msp.Identity {
	return u
}

// Sign the message
func (u *walletIdentity) Sign(msg []byte) ([]byte, error) {
	return nil, errors.New("Sign() function not implemented")
}

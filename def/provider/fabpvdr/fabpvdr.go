/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/api/apifabca"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	identityImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

// FabricProvider represents the default implementation of Fabric objects.
type FabricProvider struct {
	config      apiconfig.Config
	stateStore  apifabclient.KeyValueStore
	cryptoSuite apicryptosuite.CryptoSuite
	signer      apifabclient.SigningManager
}

// NewFabricProvider creates a FabricProvider enabling access to core Fabric objects and functionality.
func NewFabricProvider(config apiconfig.Config, stateStore apifabclient.KeyValueStore, cryptoSuite apicryptosuite.CryptoSuite, signer apifabclient.SigningManager) *FabricProvider {
	f := FabricProvider{
		config,
		stateStore,
		cryptoSuite,
		signer,
	}
	return &f
}

// NewClient returns a new FabricClient initialized for the current instance of the SDK
func (f *FabricProvider) NewClient(user apifabclient.User) (apifabclient.FabricClient, error) {
	client := clientImpl.NewClient(f.config)

	client.SetCryptoSuite(f.cryptoSuite)
	client.SetStateStore(f.stateStore)
	client.SetUserContext(user)
	client.SetSigningManager(f.signer)

	return client, nil
}

// NewCAClient returns a new FabricCAClient initialized for the current instance of the SDK
func (f *FabricProvider) NewCAClient(orgID string) (apifabca.FabricCAClient, error) {
	return fabricCAClient.NewFabricCAClient(orgID, f.config, f.cryptoSuite)
}

// NewUser returns a new default implementation of a User.
func (f *FabricProvider) NewUser(name string, signingIdentity *apifabclient.SigningIdentity) (apifabclient.User, error) {

	user := identityImpl.NewUser(name, signingIdentity.MspID)

	user.SetPrivateKey(signingIdentity.PrivateKey)
	user.SetEnrollmentCertificate(signingIdentity.EnrollmentCert)

	return user, nil
}

// NewPeer returns a new default implementation of Peer
func (f *FabricProvider) NewPeer(url string, certificate string, serverHostOverride string) (apifabclient.Peer, error) {
	return peerImpl.NewPeerTLSFromCert(url, certificate, serverHostOverride, f.config)
}

// NewPeerFromConfig returns a new default implementation of Peer based configuration
func (f *FabricProvider) NewPeerFromConfig(peerCfg *apiconfig.NetworkPeer) (apifabclient.Peer, error) {
	return peerImpl.NewPeerFromConfig(peerCfg, f.config)
}

/*
TODO: Unclear that this EnrollUser helper is really needed at this level - I think not.
Note: I renamed NewPreEnrolledUser to NewUser; and the old NewUser to EnrollUser
// NewUser returns a new default implementation of a User.
func (f *FabricProvider) EnrollUser(orgID, name, pwd string) (apifabca.User, error) {
	mspID, err := f.config.MspID(orgID)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading MSP ID config")
	}

	msp, err := f.NewCAClient(orgID)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading MSP ID config")
	}

	key, cert, err := msp.Enroll(name, pwd)
	if err != nil {
		return nil, errors.WithMessage(err, "Enroll failed")
	}
	user := identityImpl.NewUser(name, mspID)
	user.SetPrivateKey(key)
	user.SetEnrollmentCertificate(cert)

	return user, nil
}
*/

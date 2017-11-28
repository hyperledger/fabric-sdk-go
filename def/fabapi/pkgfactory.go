/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabapi

import (
	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	fabca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/config"
	cryptosuite "github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	eventsImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	identityImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	ordererImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/signingmgr"
)

// TODO: Some (or most) of these functions should no longer be exported, as usage should occur via NewSDK

// NewClient returns a new default implementation of the Client interface using the config provided.
// It will save the provided user if requested into the state store.
func NewClient(user fabca.User, skipUserPersistence bool, stateStorePath string, cryptosuiteprovider apicryptosuite.CryptoSuite, config config.Config) (fab.FabricClient, error) {
	client := clientImpl.NewClient(config)

	client.SetCryptoSuite(cryptosuiteprovider)
	if stateStorePath != "" {
		stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
		if err != nil {
			return nil, errors.WithMessage(err, "CreateNewFileKeyValueStore failed")
		}
		client.SetStateStore(stateStore)
	}
	client.SaveUserToStateStore(user, skipUserPersistence)

	signingMgr, err := signingmgr.NewSigningManager(cryptosuiteprovider, config)
	if err != nil {
		return nil, errors.WithMessage(err, "NewSigningManager failed")
	}

	client.SetSigningManager(signingMgr)

	return client, nil
}

// NewClientWithUser returns a new default implementation of the Client interface.
// It creates a default implementation of User, enrolls the user, and saves it to the state store.
func NewClientWithUser(name string, pwd string, orgName string,
	stateStorePath string, cryptosuiteprovider apicryptosuite.CryptoSuite, config config.Config, msp fabca.FabricCAClient) (fab.FabricClient, error) {
	client := clientImpl.NewClient(config)

	client.SetCryptoSuite(cryptosuiteprovider)
	stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
	if err != nil {
		return nil, errors.WithMessage(err, "CreateNewFileKeyValueStore failed")
	}
	client.SetStateStore(stateStore)
	mspID, err := client.Config().MspID(orgName)
	if err != nil {
		return nil, errors.WithMessage(err, "reading MSP ID config failed")
	}

	user, err := NewUser(client.Config(), msp, name, pwd, mspID)
	if err != nil {
		return nil, errors.WithMessage(err, "NewUser failed")
	}
	err = client.SaveUserToStateStore(user, false)
	if err != nil {
		return nil, errors.WithMessage(err, "SaveUserToStateStore failed")
	}

	client.SetUserContext(user)

	return client, nil
}

// NewUser returns a new default implementation of a User.
func NewUser(config config.Config, msp fabca.FabricCAClient, name string, pwd string,
	mspID string) (fabca.User, error) {

	key, cert, err := msp.Enroll(name, pwd)
	if err != nil {
		return nil, errors.WithMessage(err, "Enroll failed")
	}
	user := identityImpl.NewUser(name, mspID)
	user.SetPrivateKey(key)
	user.SetEnrollmentCertificate(cert)

	return user, nil
}

// NewPreEnrolledUser returns a new default implementation of a User.
func NewPreEnrolledUser(config config.Config, name string, signingIdentity *fab.SigningIdentity) (fabca.User, error) {

	user := identityImpl.NewUser(name, signingIdentity.MspID)

	user.SetPrivateKey(signingIdentity.PrivateKey)
	user.SetEnrollmentCertificate(signingIdentity.EnrollmentCert)

	return user, nil
}

// NewChannel returns a new default implementation of Channel
func NewChannel(client fab.FabricClient, orderer fab.Orderer, peers []fab.Peer, channelID string) (fab.Channel, error) {

	channel, err := client.NewChannel(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "NewChannel failed")
	}

	err = channel.AddOrderer(orderer)
	if err != nil {
		return nil, errors.WithMessage(err, "AddOrderer failed")
	}

	for _, p := range peers {
		err = channel.AddPeer(p)
		if err != nil {
			return nil, errors.WithMessage(err, "adding peer failed")
		}
	}

	return channel, nil
}

// NewSystemClient returns a new default implementation of Client
func NewSystemClient(config config.Config) *clientImpl.Client {
	return clientImpl.NewClient(config)
}

// NewKVStore returns a new default implementation of State Store
func NewKVStore(stateStorePath string) (fab.KeyValueStore, error) {
	stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
	if err != nil {
		return nil, errors.WithMessage(err, "CreateNewFileKeyValueStore failed")
	}
	return stateStore, nil
}

// NewCryptoSuite returns a new default implementation of CryptoSuite
func NewCryptoSuite(config config.Config) (apicryptosuite.CryptoSuite, error) {
	return cryptosuite.GetSuiteByConfig(config)
}

// NewSigningManager returns a new default implementation of signing manager
func NewSigningManager(cryptoProvider apicryptosuite.CryptoSuite, config config.Config) (fab.SigningManager, error) {
	return signingmgr.NewSigningManager(cryptoProvider, config)
}

// NewEventHub returns a new default implementation of EventHub
func NewEventHub(client fab.FabricClient) (fab.EventHub, error) {
	return eventsImpl.NewEventHub(client)
}

// NewOrderer returns a new default implementation of Orderer
func NewOrderer(url string, certificate string, serverHostOverride string, config config.Config) (fab.Orderer, error) {
	return ordererImpl.NewOrderer(url, certificate, serverHostOverride, config)
}

// NewOrdererFromConfig returns a new default implementation of Orderer based on config
func NewOrdererFromConfig(ordererCfg config.OrdererConfig, config config.Config) (fab.Orderer, error) {
	return ordererImpl.NewOrdererFromConfig(&ordererCfg, config)
}

// NewPeer returns a new default implementation of Peer
func NewPeer(url string, certificate string, serverHostOverride string, config config.Config) (fab.Peer, error) {
	return peerImpl.NewPeerTLSFromCert(url, certificate, serverHostOverride, config)
}

// NewPeerFromConfig returns a new default implementation of Peer based configuration
func NewPeerFromConfig(peerCfg *config.NetworkPeer, config config.Config) (fab.Peer, error) {
	return peerImpl.NewPeerFromConfig(peerCfg, config)
}

// NewConfigManager returns a new default implementation of the Config interface
func NewConfigManager(configFile string) (config.Config, error) {
	return configImpl.InitConfig(configFile)
}

// NewCAClient returns a new default implmentation of the MSP client
func NewCAClient(orgName string, config config.Config, cryptoSuite apicryptosuite.CryptoSuite) (fabca.FabricCAClient, error) {
	mspClient, err := fabricCAClient.NewFabricCAClient(orgName, config, cryptoSuite)
	if err != nil {
		return nil, errors.WithMessage(err, "NewFabricCAClient failed")
	}

	return mspClient, nil
}

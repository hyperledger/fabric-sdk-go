/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabapi

import (
	"fmt"
	"io/ioutil"

	fabricCaUtil "github.com/hyperledger/fabric-ca/util"
	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fabca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/config"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	eventsImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	identityImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	ordererImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	bccsp "github.com/hyperledger/fabric/bccsp"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
)

// TODO: Some (or most) of these functions should no longer be exported, as usage should occur via NewSDK

// NewClient returns a new default implementation of the Client interface using the config provided.
// It will save the provided user if requested into the state store.
func NewClient(user fabca.User, skipUserPersistence bool, stateStorePath string, cryptosuite bccsp.BCCSP, config config.Config) (fab.FabricClient, error) {
	client := clientImpl.NewClient(config)

	if cryptosuite == nil {
		cryptosuite = bccspFactory.GetDefault()
	}
	client.SetCryptoSuite(cryptosuite)
	if stateStorePath != "" {
		stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
		if err != nil {
			return nil, fmt.Errorf("CreateNewFileKeyValueStore returned error[%s]", err)
		}
		client.SetStateStore(stateStore)
	}
	client.SaveUserToStateStore(user, skipUserPersistence)

	return client, nil
}

// NewClientWithUser returns a new default implementation of the Client interface.
// It creates a default implementation of User, enrolls the user, and saves it to the state store.
func NewClientWithUser(name string, pwd string, orgName string,
	stateStorePath string, cryptosuite bccsp.BCCSP, config config.Config, msp fabca.FabricCAClient) (fab.FabricClient, error) {
	client := clientImpl.NewClient(config)

	if cryptosuite == nil {
		cryptosuite = bccspFactory.GetDefault()
	}
	client.SetCryptoSuite(cryptosuite)
	stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
	if err != nil {
		return nil, fmt.Errorf("CreateNewFileKeyValueStore returned error[%s]", err)
	}
	client.SetStateStore(stateStore)
	mspID, err := client.Config().MspID(orgName)
	if err != nil {
		return nil, fmt.Errorf("Error reading MSP ID config: %s", err)
	}

	user, err := NewUser(client.Config(), msp, name, pwd, mspID)
	if err != nil {
		return nil, fmt.Errorf("NewUser returned error: %v", err)
	}
	err = client.SaveUserToStateStore(user, false)
	if err != nil {
		return nil, fmt.Errorf("client.SaveUserToStateStore returned error: %v", err)
	}

	client.SetUserContext(user)

	return client, nil
}

// NewClientWithPreEnrolledUser returns a new default Client implementation
// by using a the default implementation of a pre-enrolled user.
func NewClientWithPreEnrolledUser(config config.Config, stateStorePath string,
	skipUserPersistence bool, username string, keyDir string, certDir string,
	orgName string, cryptosuite bccsp.BCCSP) (fab.FabricClient, error) {

	client := clientImpl.NewClient(config)

	if cryptosuite == nil {
		cryptosuite = bccspFactory.GetDefault()
	}
	client.SetCryptoSuite(cryptosuite)
	if stateStorePath != "" {
		stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
		if err != nil {
			return nil, fmt.Errorf("CreateNewFileKeyValueStore returned error[%s]", err)
		}
		client.SetStateStore(stateStore)
	}
	mspID, err := client.Config().MspID(orgName)
	if err != nil {
		return nil, fmt.Errorf("Error reading MSP ID config: %s", err)
	}
	user, err := NewPreEnrolledUser(client.Config(), keyDir, certDir, username, mspID, client.CryptoSuite())
	if err != nil {
		return nil, fmt.Errorf("NewPreEnrolledUser returned error: %v", err)
	}
	client.SetUserContext(user)
	client.SaveUserToStateStore(user, skipUserPersistence)

	return client, nil
}

// NewUser returns a new default implementation of a User.
func NewUser(config config.Config, msp fabca.FabricCAClient, name string, pwd string,
	mspID string) (fabca.User, error) {

	key, cert, err := msp.Enroll(name, pwd)
	if err != nil {
		return nil, fmt.Errorf("Enroll returned error: %v", err)
	}
	user := identityImpl.NewUser(name, mspID)
	user.SetPrivateKey(key)
	user.SetEnrollmentCertificate(cert)

	return user, nil
}

// NewPreEnrolledUser returns a new default implementation of User.
// The user should already be pre-enrolled.
func NewPreEnrolledUser(config config.Config, privateKeyPath string,
	enrollmentCertPath string, username string, mspID string, cryptoSuite bccsp.BCCSP) (fabca.User, error) {
	privateKey, err := fabricCaUtil.ImportBCCSPKeyFromPEM(privateKeyPath, cryptoSuite, true)
	if err != nil {
		return nil, fmt.Errorf("Error importing private key: %v", err)
	}
	enrollmentCert, err := ioutil.ReadFile(enrollmentCertPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading from the enrollment cert path: %v", err)
	}

	user := identityImpl.NewUser(username, mspID)
	user.SetEnrollmentCertificate(enrollmentCert)
	user.SetPrivateKey(privateKey)

	return user, nil
}

// NewChannel returns a new default implementation of Channel
func NewChannel(client fab.FabricClient, orderer fab.Orderer, peers []fab.Peer, channelID string) (fab.Channel, error) {

	channel, err := client.NewChannel(channelID)
	if err != nil {
		return nil, fmt.Errorf("NewChannel returned error: %v", err)
	}

	err = channel.AddOrderer(orderer)
	if err != nil {
		return nil, fmt.Errorf("Error adding orderer: %v", err)
	}

	for _, p := range peers {
		err = channel.AddPeer(p)
		if err != nil {
			return nil, fmt.Errorf("Error adding peer: %v", err)
		}
	}

	return channel, nil
}

// NewSystemClient returns a new default implementation of Client
func NewSystemClient(config config.Config) fab.FabricClient {
	return clientImpl.NewClient(config)
}

// NewKVStore returns a new default implementation of State Store
func NewKVStore(stateStorePath string) (fab.KeyValueStore, error) {
	stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
	if err != nil {
		return nil, fmt.Errorf("CreateNewFileKeyValueStore returned error[%s]", err)
	}
	return stateStore, nil
}

// NewCryptoSuite returns a new default implementation of BCCSP
func NewCryptoSuite(config *bccspFactory.FactoryOpts) (bccsp.BCCSP, error) {
	return bccspFactory.GetBCCSPFromOpts(config)
}

// NewEventHub returns a new default implementation of EventHub
func NewEventHub(client fab.FabricClient) (fab.EventHub, error) {
	return eventsImpl.NewEventHub(client)
}

// NewOrderer returns a new default implementation of Orderer
func NewOrderer(url string, certificate string, serverHostOverride string, config config.Config) (fab.Orderer, error) {
	return ordererImpl.NewOrderer(url, certificate, serverHostOverride, config)
}

// NewPeer returns a new default implementation of Peer
func NewPeer(url string, certificate string, serverHostOverride string, config config.Config) (fab.Peer, error) {
	return peerImpl.NewPeerTLSFromCert(url, certificate, serverHostOverride, config)
}

// NewConfig returns a new default implementation of the Config interface
func NewConfigManager(configFile string) (config.Config, error) {
	return configImpl.InitConfig(configFile)
}

// NewCAClient returns a new default implmentation of the MSP client
func NewCAClient(orgName string, config config.Config) (fabca.FabricCAClient, error) {
	mspClient, err := fabricCAClient.NewFabricCAClient(config, orgName)
	if err != nil {
		return nil, fmt.Errorf("NewFabricCAClient returned error: %v", err)
	}

	return mspClient, nil
}

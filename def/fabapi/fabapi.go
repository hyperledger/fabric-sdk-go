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
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	mspImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/msp"
	ordererImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	bccsp "github.com/hyperledger/fabric/bccsp"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
)

// NewClient returns a new default implementation of the Client interface using the config provided.
// It will save the provided user if requested into the state store.
func NewClient(user fabca.User, skipUserPersistence bool, stateStorePath string, config config.Config) (fab.FabricClient, error) {
	client := clientImpl.NewClient(config)

	cryptoSuite := bccspFactory.GetDefault()

	client.SetCryptoSuite(cryptoSuite)
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
	stateStorePath string, config config.Config, msp fabca.FabricCAClient) (fab.FabricClient, error) {
	client := clientImpl.NewClient(config)

	cryptoSuite := bccspFactory.GetDefault()

	client.SetCryptoSuite(cryptoSuite)
	stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
	if err != nil {
		return nil, fmt.Errorf("CreateNewFileKeyValueStore returned error[%s]", err)
	}
	client.SetStateStore(stateStore)
	mspID, err := client.GetConfig().MspID(orgName)
	if err != nil {
		return nil, fmt.Errorf("Error reading MSP ID config: %s", err)
	}

	user, err := NewUser(client.GetConfig(), msp, name, pwd, mspID)
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
	orgName string) (fab.FabricClient, error) {

	client := clientImpl.NewClient(config)

	cryptoSuite := bccspFactory.GetDefault()

	client.SetCryptoSuite(cryptoSuite)
	if stateStorePath != "" {
		stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
		if err != nil {
			return nil, fmt.Errorf("CreateNewFileKeyValueStore returned error[%s]", err)
		}
		client.SetStateStore(stateStore)
	}
	mspID, err := client.GetConfig().MspID(orgName)
	if err != nil {
		return nil, fmt.Errorf("Error reading MSP ID config: %s", err)
	}
	user, err := NewPreEnrolledUser(client.GetConfig(), keyDir, certDir, username, mspID, client.GetCryptoSuite())
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
	user := mspImpl.NewUser(name, mspID)
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

	user := mspImpl.NewUser(username, mspID)
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

// NewEventHub returns a new default implementation of Event Hub
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
func NewConfig(configFile string) (config.Config, error) {
	return configImpl.InitConfig(configFile)
}

// NewCAClient returns a new default implmentation of the MSP client
func NewCAClient(config config.Config, orgName string) (fabca.FabricCAClient, error) {
	mspClient, err := fabricCAClient.NewFabricCAClient(config, orgName)
	if err != nil {
		return nil, fmt.Errorf("NewFabricCAClient returned error: %v", err)
	}

	return mspClient, nil
}

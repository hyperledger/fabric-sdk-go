/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defaultimpl

import (
	"fmt"
	"io/ioutil"

	fabricCaUtil "github.com/hyperledger/fabric-ca/util"
	api "github.com/hyperledger/fabric-sdk-go/api"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/config"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	eventsImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	ordererImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	userImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/user"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
)

// NewClient returns a new default implementation of the Client interface using the config provided.
// It will save the provided user if requested into the state store.
func NewClient(user api.User, skipUserPersistence bool, stateStorePath string, config api.Config) (api.FabricClient, error) {
	client := clientImpl.NewClient(config)

	cryptoSuite := bccspFactory.GetDefault()

	client.SetCryptoSuite(cryptoSuite)
	stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
	if err != nil {
		return nil, fmt.Errorf("CreateNewFileKeyValueStore returned error[%s]", err)
	}
	client.SetStateStore(stateStore)
	client.SaveUserToStateStore(user, skipUserPersistence)

	return client, nil
}

// NewClientWithUser returns a new default implementation of the Client interface.
// It creates a default implementation of User, enrolls the user, and saves it to the state store.
func NewClientWithUser(name string, pwd string, orgName string,
	stateStorePath string, config api.Config, msp api.FabricCAClient) (api.FabricClient, error) {
	client := clientImpl.NewClient(config)

	cryptoSuite := bccspFactory.GetDefault()

	client.SetCryptoSuite(cryptoSuite)
	stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
	if err != nil {
		return nil, fmt.Errorf("CreateNewFileKeyValueStore returned error[%s]", err)
	}
	client.SetStateStore(stateStore)

	user, err := NewUser(client, msp, name, pwd, orgName)
	if err != nil {
		return nil, fmt.Errorf("NewUser returned error: %v", err)
	}
	client.SetUserContext(user)

	return client, nil
}

// NewClientWithPreEnrolledUser returns a new default Client implementation
// by using a the default implementation of a pre-enrolled user.
func NewClientWithPreEnrolledUser(config api.Config, stateStorePath string,
	skipUserPersistence bool, username string, keyDir string, certDir string,
	orgName string) (api.FabricClient, error) {

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
	user, err := NewPreEnrolledUser(client, keyDir, certDir, username, orgName)
	if err != nil {
		return nil, fmt.Errorf("NewPreEnrolledUser returned error: %v", err)
	}
	client.SetUserContext(user)
	client.SaveUserToStateStore(user, skipUserPersistence)

	return client, nil
}

// NewUser returns a new default implementation of a User.
func NewUser(client api.FabricClient, msp api.FabricCAClient, name string, pwd string,
	orgName string) (api.User, error) {
	user, err := client.LoadUserFromStateStore(name)
	if err != nil {
		return nil, fmt.Errorf("client.LoadUserFromStateStore returned error: %v", err)
	}

	if user == nil {
		mspID, err := client.GetConfig().GetMspID(orgName)
		if err != nil {
			return nil, fmt.Errorf("Error reading MSP ID config: %s", err)
		}

		key, cert, err := msp.Enroll(name, pwd)
		if err != nil {
			return nil, fmt.Errorf("Enroll returned error: %v", err)
		}
		user = userImpl.NewUser(name, mspID)
		user.SetPrivateKey(key)
		user.SetEnrollmentCertificate(cert)
		err = client.SaveUserToStateStore(user, false)
		if err != nil {
			return nil, fmt.Errorf("client.SaveUserToStateStore returned error: %v", err)
		}
	}

	return user, nil
}

// NewPreEnrolledUser returns a new default implementation of User.
// The user should already be pre-enrolled.
func NewPreEnrolledUser(client api.FabricClient, privateKeyPath string,
	enrollmentCertPath string, username string, orgName string) (api.User, error) {
	mspID, err := client.GetConfig().GetMspID(orgName)
	if err != nil {
		return nil, fmt.Errorf("Error reading MSP ID config: %s", err)
	}
	privateKey, err := fabricCaUtil.ImportBCCSPKeyFromPEM(privateKeyPath, client.GetCryptoSuite(), true)
	if err != nil {
		return nil, fmt.Errorf("Error importing private key: %v", err)
	}
	enrollmentCert, err := ioutil.ReadFile(enrollmentCertPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading from the enrollment cert path: %v", err)
	}

	user := userImpl.NewUser(username, mspID)
	user.SetEnrollmentCertificate(enrollmentCert)
	user.SetPrivateKey(privateKey)

	return user, nil
}

// NewChannel returns a new default implementation of Channel
func NewChannel(client api.FabricClient, orderer api.Orderer, peers []api.Peer, channelID string) (api.Channel, error) {

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
func NewEventHub(client api.FabricClient) (api.EventHub, error) {
	return eventsImpl.NewEventHub(client)
}

// NewOrderer returns a new default implementation of Orderer
func NewOrderer(url string, certificate string, serverHostOverride string, config api.Config) (api.Orderer, error) {
	return ordererImpl.NewOrderer(url, certificate, serverHostOverride, config)
}

// NewPeer returns a new default implementation of Peer
func NewPeer(url string, certificate string, serverHostOverride string, config api.Config) (api.Peer, error) {
	return peerImpl.NewPeerTLSFromCert(url, certificate, serverHostOverride, config)
}

// NewConfig returns a new default implementation of the Config interface
func NewConfig(configFile string) (api.Config, error) {
	return configImpl.InitConfig(configFile)
}

// NewCAClient returns a new default implmentation of the MSP client
func NewCAClient(config api.Config, orgName string) (api.FabricCAClient, error) {
	mspClient, err := fabricCAClient.NewFabricCAClient(config, orgName)
	if err != nil {
		return nil, fmt.Errorf("NewFabricCAClient returned error: %v", err)
	}

	return mspClient, nil
}

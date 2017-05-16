/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fabricclient

import (
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/golang/protobuf/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-sdk-go/config"
	kvs "github.com/hyperledger/fabric-sdk-go/fabric-client/keyvaluestore"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric/protos/utils"
)

// Client ...
/*
 * Main interaction handler with end user. A client instance provides a handler to interact
 * with a network of peers, orderers and optionally member services. An application using the
 * SDK may need to interact with multiple networks, each through a separate instance of the Client.
 *
 * Each client when initially created should be initialized with configuration data from the
 * consensus service, which includes a list of trusted roots, orderer certificates and IP addresses,
 * and a list of peer certificates and IP addresses that it can access. This must be done out of band
 * as part of bootstrapping the application environment. It is also the responsibility of the application
 * to maintain the configuration of a client as the SDK does not persist this object.
 *
 * Each Client instance can maintain several {@link Chain} instances representing channels and the associated
 * private ledgers.
 *
 *
 */
type Client interface {
	NewChain(name string) (Chain, error)
	GetChain(name string) Chain
	CreateChannel(request *CreateChannelRequest) (Chain, error)
	QueryChainInfo(name string, peers []Peer) (Chain, error)
	SetStateStore(stateStore kvs.KeyValueStore)
	GetStateStore() kvs.KeyValueStore
	SetCryptoSuite(cryptoSuite bccsp.BCCSP)
	GetCryptoSuite() bccsp.BCCSP
	SaveUserToStateStore(user User, skipPersistence bool) error
	LoadUserFromStateStore(name string) (User, error)
	InstallChaincode(chaincodeName string, chaincodePath string, chaincodeVersion string, chaincodePackage []byte, targets []Peer) ([]*TransactionProposalResponse, string, error)
	QueryChannels(peer Peer) (*pb.ChannelQueryResponse, error)
	QueryInstalledChaincodes(peer Peer) (*pb.ChaincodeQueryResponse, error)
}

type client struct {
	chains      map[string]Chain
	cryptoSuite bccsp.BCCSP
	stateStore  kvs.KeyValueStore
	userContext User
}

// CreateChannelRequest requests channel creation on the network
type CreateChannelRequest struct {
	// The name of the channel
	Name string
	// Orderer object to create the channel
	Orderer Orderer
	// Contains channel configuration (ConfigTx)
	Envelope []byte
}

// NewClient ...
/*
 * Returns a Client instance
 */
func NewClient() Client {
	chains := make(map[string]Chain)
	c := &client{chains: chains, cryptoSuite: nil, stateStore: nil, userContext: nil}
	return c
}

// NewChain ...
/*
 * Returns a chain instance with the given name. This represents a channel and its associated ledger
 * (as explained above), and this call returns an empty object. To initialize the chain in the blockchain network,
 * a list of participating endorsers and orderer peers must be configured first on the returned object.
 * @param {string} name The name of the chain.  Recommend using namespaces to avoid collision.
 * @returns {Chain} The uninitialized chain instance.
 * @returns {Error} if the chain by that name already exists in the application's state store
 */
func (c *client) NewChain(name string) (Chain, error) {
	if _, ok := c.chains[name]; ok {
		return nil, fmt.Errorf("Chain %s already exists", name)
	}
	var err error
	c.chains[name], err = NewChain(name, c)
	if err != nil {
		return nil, err
	}
	return c.chains[name], nil
}

// GetChain ...
/*
 * Get a {@link Chain} instance from the state storage. This allows existing chain instances to be saved
 * for retrieval later and to be shared among instances of the application. Note that it’s the
 * application/SDK’s responsibility to record the chain information. If an application is not able
 * to look up the chain information from storage, it may call another API that queries one or more
 * Peers for that information.
 * @param {string} name The name of the chain.
 * @returns {Chain} The chain instance
 */
func (c *client) GetChain(name string) Chain {
	return c.chains[name]
}

// QueryChainInfo ...
/*
 * This is a network call to the designated Peer(s) to discover the chain information.
 * The target Peer(s) must be part of the chain to be able to return the requested information.
 * @param {string} name The name of the chain.
 * @param {[]Peer} peers Array of target Peers to query.
 * @returns {Chain} The chain instance for the name or error if the target Peer(s) does not know
 * anything about the chain.
 */
func (c *client) QueryChainInfo(name string, peers []Peer) (Chain, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

// SetStateStore ...
/*
 * The SDK should have a built-in key value store implementation (suggest a file-based implementation to allow easy setup during
 * development). But production systems would want a store backed by database for more robust storage and clustering,
 * so that multiple app instances can share app state via the database (note that this doesn’t necessarily make the app stateful).
 * This API makes this pluggable so that different store implementations can be selected by the application.
 */
func (c *client) SetStateStore(stateStore kvs.KeyValueStore) {
	c.stateStore = stateStore
}

// GetStateStore ...
/*
 * A convenience method for obtaining the state store object in use for this client.
 */
func (c *client) GetStateStore() kvs.KeyValueStore {
	return c.stateStore
}

// SetCryptoSuite ...
/*
 * A convenience method for obtaining the state store object in use for this client.
 */
func (c *client) SetCryptoSuite(cryptoSuite bccsp.BCCSP) {
	c.cryptoSuite = cryptoSuite
}

// GetCryptoSuite ...
/*
 * A convenience method for obtaining the CryptoSuite object in use for this client.
 */
func (c *client) GetCryptoSuite() bccsp.BCCSP {
	return c.cryptoSuite
}

// SaveUserToStateStore ...
/*
 * Sets an instance of the User class as the security context of this client instance. This user’s credentials (ECert) will be
 * used to conduct transactions and queries with the blockchain network. Upon setting the user context, the SDK saves the object
 * in a persistence cache if the “state store” has been set on the Client instance. If no state store has been set,
 * this cache will not be established and the application is responsible for setting the user context again when the application
 * crashed and is recovered.
 */
func (c *client) SaveUserToStateStore(user User, skipPersistence bool) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	if user.GetName() == "" {
		return fmt.Errorf("user name is empty")
	}
	c.userContext = user
	if !skipPersistence {
		if c.stateStore == nil {
			return fmt.Errorf("stateStore is nil")
		}
		userJSON := &UserJSON{PrivateKeySKI: user.GetPrivateKey().SKI(), EnrollmentCertificate: user.GetEnrollmentCertificate()}
		data, err := json.Marshal(userJSON)
		if err != nil {
			return fmt.Errorf("Marshal json return error: %v", err)
		}
		err = c.stateStore.SetValue(user.GetName(), data)
		if err != nil {
			return fmt.Errorf("stateStore SaveUserToStateStore return error: %v", err)
		}
	}
	return nil

}

// LoadUserFromStateStore ...
/*
 * The client instance can have an optional state store. The SDK saves enrolled users in the storage which can be accessed by
 * authorized users of the application (authentication is done by the application outside of the SDK).
 * This function attempts to load the user by name from the local storage (via the KeyValueStore interface).
 * The loaded user object must represent an enrolled user with a valid enrollment certificate signed by a trusted CA
 * (such as the COP server).
 */
func (c *client) LoadUserFromStateStore(name string) (User, error) {
	if c.userContext != nil {
		return c.userContext, nil
	}
	if name == "" {
		return nil, nil
	}
	if c.stateStore == nil {
		return nil, nil
	}
	if c.cryptoSuite == nil {
		return nil, fmt.Errorf("cryptoSuite is nil")
	}
	value, err := c.stateStore.GetValue(name)
	if err != nil {
		return nil, nil
	}
	var userJSON UserJSON
	err = json.Unmarshal(value, &userJSON)
	if err != nil {
		return nil, fmt.Errorf("stateStore GetValue return error: %v", err)
	}
	user := NewUser(name)
	user.SetEnrollmentCertificate(userJSON.EnrollmentCertificate)
	key, err := c.cryptoSuite.GetKey(userJSON.PrivateKeySKI)
	if err != nil {
		return nil, fmt.Errorf("cryptoSuite GetKey return error: %v", err)
	}
	user.SetPrivateKey(key)
	c.userContext = user
	return c.userContext, nil

}

// CreateChannel calls an orderer to create a channel on the network.
// Only one of the application instances needs to call this method.
// Once the chain is successfully created, this and other application
// instances only need to call Chain joinChannel() to participate on the channel
func (c *client) CreateChannel(request *CreateChannelRequest) (Chain, error) {
	// Validate request
	if request == nil {
		return nil, fmt.Errorf("Missing all required input request parameters for initialize channel")
	}

	if request.Envelope == nil {
		return nil, fmt.Errorf("Missing envelope request parameter containing the configuration of the new channel")
	}

	if request.Orderer == nil {
		return nil, fmt.Errorf("Missing orderer request parameter for the initialize channel")
	}

	if request.Name == "" {
		return nil, fmt.Errorf("Missing name request parameter for the new channel")
	}

	signedEnvelope := &common.Envelope{}
	err := proto.Unmarshal(request.Envelope, signedEnvelope)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling channel configuration data: %s",
			err.Error())
	}
	// Send request
	err = request.Orderer.SendBroadcast(&SignedEnvelope{
		Signature: signedEnvelope.Signature,
		Payload:   signedEnvelope.Payload,
	})
	if err != nil {
		return nil, fmt.Errorf("Could not broadcast to orderer %s: %s", request.Orderer.GetURL(), err.Error())
	}

	// Initialize the new chain

	// FIXME: Temporary code checks if the chain already exists
	// and, if not, creates the chain. This check should be removed after
	// the end-to-end test is refactored to not create the chain before invoking CreateChannel
	chain := c.GetChain(request.Name)
	if chain == nil {
		logger.Debugf("Creating new chain: %", request.Name)
		chain, err = c.NewChain(request.Name)
		if err != nil {
			return nil, fmt.Errorf("Error while creating new chain %s: %v", request.Name, err)
		}
	}

	if err := chain.Initialize(request.Envelope); err != nil {
		return nil, fmt.Errorf("Error while initializing chain: %v", err)
	}

	chain.AddOrderer(request.Orderer)

	return chain, nil
}

// NewPeer ...
/**
 * Constructs a Peer given its endpoint configuration settings.
 *
 * @param {string} url The URL with format of "host:port".
 */
func NewPeer(url string, certificate string, serverHostOverride string) (Peer, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(time.Second*3))
	if config.IsTLSEnabled() {
		tlsCaCertPool, err := config.GetTLSCACertPool(certificate)
		if err != nil {
			return nil, err
		}
		creds := credentials.NewClientTLSFromCert(tlsCaCertPool, serverHostOverride)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	return &peer{url: url, grpcDialOption: opts, name: "", roles: nil}, nil

}

// NewOrderer Returns a Orderer instance
func NewOrderer(url string, certificate string, serverHostOverride string) (Orderer, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(time.Second*3))
	if config.IsTLSEnabled() {
		tlsCaCertPool, err := config.GetTLSCACertPool(certificate)
		if err != nil {
			return nil, err
		}
		creds := credentials.NewClientTLSFromCert(tlsCaCertPool, serverHostOverride)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	return &orderer{url: url, grpcDialOption: opts}, nil
}

//QueryChannels
/**
 * Queries the names of all the channels that a
 * peer has joined.
 * @param {Peer} peer
 * @returns {object} ChannelQueryResponse proto
 */
func (c *client) QueryChannels(peer Peer) (*pb.ChannelQueryResponse, error) {

	if peer == nil {
		return nil, fmt.Errorf("QueryChannels requires peer")
	}

	responses, err := QueryByChaincode("cscc", []string{"GetChannels"}, []Peer{peer}, c)
	if err != nil {
		return nil, fmt.Errorf("QueryByChaincode return error: %v", err)
	}

	payload := responses[0]
	response := new(pb.ChannelQueryResponse)
	err = proto.Unmarshal(payload, response)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal ChannelQueryResponse return error: %v", err)
	}
	return response, nil
}

//QueryInstalledChaincodes
/**
 * Queries the installed chaincodes on a peer
 * Returning the details of all chaincodes installed on a peer.
 * @param {Peer} peer
 * @returns {object} ChaincodeQueryResponse proto
 */
func (c *client) QueryInstalledChaincodes(peer Peer) (*pb.ChaincodeQueryResponse, error) {

	if peer == nil {
		return nil, fmt.Errorf("To query installed chaincdes you need to pass peer")
	}
	responses, err := QueryByChaincode("lccc", []string{"getinstalledchaincodes"}, []Peer{peer}, c)
	if err != nil {
		return nil, fmt.Errorf("Invoke lccc getinstalledchaincodes return error: %v", err)
	}
	payload := responses[0]
	response := new(pb.ChaincodeQueryResponse)
	err = proto.Unmarshal(payload, response)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal ChaincodeQueryResponse return error: %v", err)
	}

	return response, nil
}

// InstallChaincode
/**
* Sends an install proposal to one or more endorsing peers.
* @param {string} chaincodeName: required - The name of the chaincode.
* @param {[]string} chaincodePath: required - string of the path to the location of the source code of the chaincode
* @param {[]string} chaincodeVersion: required - string of the version of the chaincode
* @param {[]string} chaincodeVersion: optional - Array of byte the chaincodePackage
 */
func (c *client) InstallChaincode(chaincodeName string, chaincodePath string, chaincodeVersion string,
	chaincodePackage []byte, targets []Peer) ([]*TransactionProposalResponse, string, error) {

	if chaincodeName == "" {
		return nil, "", fmt.Errorf("Missing 'chaincodeName' parameter")
	}
	if chaincodePath == "" {
		return nil, "", fmt.Errorf("Missing 'chaincodePath' parameter")
	}
	if chaincodeVersion == "" {
		return nil, "", fmt.Errorf("Missing 'chaincodeVersion' parameter")
	}

	if chaincodePackage == nil {
		var err error
		chaincodePackage, err = PackageCC(chaincodePath, "")
		if err != nil {
			return nil, "", fmt.Errorf("PackageCC return error: %s", err)
		}
	}

	now := time.Now()
	cds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincodeName, Path: chaincodePath, Version: chaincodeVersion}},
		CodePackage: chaincodePackage, EffectiveDate: &google_protobuf.Timestamp{Seconds: int64(now.Second()), Nanos: int32(now.Nanosecond())}}

	user, err := c.LoadUserFromStateStore("")
	if err != nil {
		return nil, "", fmt.Errorf("GetUserContext return error: %s", err)
	}

	creatorID, err := getSerializedIdentity(user.GetEnrollmentCertificate())
	if err != nil {
		return nil, "", err
	}

	// create an install from a chaincodeDeploymentSpec
	proposal, txID, err := protos_utils.CreateInstallProposalFromCDS(cds, creatorID)
	if err != nil {
		return nil, "", fmt.Errorf("Could not create chaincode Deploy proposal, err %s", err)
	}
	proposalBytes, err := protos_utils.GetBytesProposal(proposal)
	if err != nil {
		return nil, "", err
	}
	signature, err := signObjectWithKey(proposalBytes, user.GetPrivateKey(), &bccsp.SHAOpts{}, nil, c.GetCryptoSuite())
	if err != nil {
		return nil, "", err
	}

	signedProposal, err := &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}, nil
	if err != nil {
		return nil, "", err
	}

	transactionProposalResponse, err := SendTransactionProposal(&TransactionProposal{
		SignedProposal: signedProposal,
		Proposal:       proposal,
		TransactionID:  txID,
	}, 0, targets)

	return transactionProposalResponse, txID, err
}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricclient

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/timestamp"
	api "github.com/hyperledger/fabric-sdk-go/api"
	channel "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	fc "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/packager"
	fcUser "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/user"

	"github.com/op/go-logging"

	"github.com/hyperledger/fabric/bccsp"
	fcutils "github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/msp"
	pb "github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric/protos/utils"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

type client struct {
	channels    map[string]api.Channel
	cryptoSuite bccsp.BCCSP
	stateStore  api.KeyValueStore
	userContext api.User
	config      api.Config
}

// NewClient ...
/*
 * Returns a Client instance
 */
func NewClient(config api.Config) api.FabricClient {
	channels := make(map[string]api.Channel)
	c := &client{channels: channels, cryptoSuite: nil, stateStore: nil, userContext: nil, config: config}
	return c
}

// NewChannel ...
/*
 * Returns a channel instance with the given name. This represents a channel and its associated ledger
 * (as explained above), and this call returns an empty object. To initialize the channel in the blockchain network,
 * a list of participating endorsers and orderer peers must be configured first on the returned object.
 * @param {string} name The name of the channel.  Recommend using namespaces to avoid collision.
 * @returns {Channel} The uninitialized channel instance.
 * @returns {Error} if the channel by that name already exists in the application's state store
 */
func (c *client) NewChannel(name string) (api.Channel, error) {
	if _, ok := c.channels[name]; ok {
		return nil, fmt.Errorf("Channel %s already exists", name)
	}
	var err error
	c.channels[name], err = channel.NewChannel(name, c)
	if err != nil {
		return nil, err
	}
	return c.channels[name], nil
}

// GetConfig ...
func (c *client) GetConfig() api.Config {
	return c.config
}

// GetChannel ...
/*
 * Get a {@link Channel} instance from the state storage. This allows existing channel instances to be saved
 * for retrieval later and to be shared among instances of the application. Note that it’s the
 * application/SDK’s responsibility to record the channel information. If an application is not able
 * to look up the channel information from storage, it may call another API that queries one or more
 * Peers for that information.
 * @param {string} name The name of the channel.
 * @returns {Channel} The channel instance
 */
func (c *client) GetChannel(name string) api.Channel {
	return c.channels[name]
}

// QueryChannelInfo ...
/*
 * This is a network call to the designated Peer(s) to discover the channel information.
 * The target Peer(s) must be part of the channel to be able to return the requested information.
 * @param {string} name The name of the channel.
 * @param {[]Peer} peers Array of target Peers to query.
 * @returns {Channel} The channel instance for the name or error if the target Peer(s) does not know
 * anything about the channel.
 */
func (c *client) QueryChannelInfo(name string, peers []api.Peer) (api.Channel, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

// SetStateStore ...
/*
 * The SDK should have a built-in key value store implementation (suggest a file-based implementation to allow easy setup during
 * development). But production systems would want a store backed by database for more robust storage and clustering,
 * so that multiple app instances can share app state via the database (note that this doesn’t necessarily make the app stateful).
 * This API makes this pluggable so that different store implementations can be selected by the application.
 */
func (c *client) SetStateStore(stateStore api.KeyValueStore) {
	c.stateStore = stateStore
}

// GetStateStore ...
/*
 * A convenience method for obtaining the state store object in use for this client.
 */
func (c *client) GetStateStore() api.KeyValueStore {
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
func (c *client) SaveUserToStateStore(user api.User, skipPersistence bool) error {
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
		userJSON := &fcUser.JSON{PrivateKeySKI: user.GetPrivateKey().SKI(), EnrollmentCertificate: user.GetEnrollmentCertificate()}
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
/**
 * Restore the state of this member from the key value store (if found).  If not found, do nothing.
 * @returns {Promise} A Promise for a {User} object upon successful restore, or if the user by the name
 * does not exist in the state store, returns null without rejecting the promise
 */
func (c *client) LoadUserFromStateStore(name string) (api.User, error) {
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
	var userJSON fcUser.JSON
	err = json.Unmarshal(value, &userJSON)
	if err != nil {
		return nil, fmt.Errorf("stateStore GetValue return error: %v", err)
	}
	user := fcUser.NewUser(name)
	user.SetEnrollmentCertificate(userJSON.EnrollmentCertificate)
	key, err := c.cryptoSuite.GetKey(userJSON.PrivateKeySKI)
	if err != nil {
		return nil, fmt.Errorf("cryptoSuite GetKey return error: %v", err)
	}
	user.SetPrivateKey(key)
	c.userContext = user
	return c.userContext, nil
}

// ExtractChannelConfig ...
/**
 * Extracts the protobuf 'ConfigUpdate' object out of the 'ConfigEnvelope'
 * that is produced by the ConfigTX tool. The returned object may then be
 * signed using the signChannelConfig() method of this class. Once the all
 * signatures have been collected this object and the signatures may be used
 * on the updateChannel or createChannel requests.
 * @param {byte[]} The bytes of the ConfigEnvelope protopuf
 * @returns {byte[]} The bytes of the ConfigUpdate protobuf
 */
func (c *client) ExtractChannelConfig(configEnvelope []byte) ([]byte, error) {
	logger.Debug("extractConfigUpdate - start")

	envelope := &common.Envelope{}
	err := proto.Unmarshal(configEnvelope, envelope)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling config envelope: %v", err)
	}

	payload := &common.Payload{}
	err = proto.Unmarshal(envelope.Payload, payload)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling config payload: %v", err)
	}

	configUpdateEnvelope := &common.ConfigUpdateEnvelope{}
	err = proto.Unmarshal(payload.Data, configUpdateEnvelope)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling config update envelope: %v", err)
	}

	return configUpdateEnvelope.ConfigUpdate, nil
}

// SignChannelConfig ...
/**
 * Sign a configuration
 * @param {byte[]} config - The Configuration Update in byte form
 * @return {ConfigSignature} - The signature of the current user on the config bytes
 */
func (c *client) SignChannelConfig(config []byte) (*common.ConfigSignature, error) {
	logger.Debug("SignChannelConfig - start")

	if config == nil {
		return nil, fmt.Errorf("Channel configuration parameter is required")
	}

	creator, err := c.GetIdentity()
	if err != nil {
		return nil, fmt.Errorf("Error getting creator: %v", err)
	}
	nonce, err := fc.GenerateRandomNonce()
	if err != nil {
		return nil, fmt.Errorf("Error generating nonce: %v", err)
	}

	// signature is across a signature header and the config update
	signatureHeader := &common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
	}
	signatureHeaderBytes, err := proto.Marshal(signatureHeader)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling signatureHeader: %v", err)
	}

	user, err := c.LoadUserFromStateStore("")
	if err != nil {
		return nil, fmt.Errorf("Error getting user from store: %s", err)
	}

	// get all the bytes to be signed together, then sign
	signingBytes := fcutils.ConcatenateBytes(signatureHeaderBytes, config)
	signature, err := fc.SignObjectWithKey(signingBytes, user.GetPrivateKey(), &bccsp.SHAOpts{}, nil, c.GetCryptoSuite())
	if err != nil {
		return nil, fmt.Errorf("error singing config: %v", err)
	}

	// build the return object
	configSignature := &common.ConfigSignature{
		SignatureHeader: signatureHeaderBytes,
		Signature:       signature,
	}
	return configSignature, nil
}

// CreateChannel ...
/**
 * Calls the orderer to start building the new channel.
 * Only one of the application instances needs to call this method.
 * Once the channel is successfully created, this and other application
 * instances only need to call Channel joinChannel() to participate on the channel.
 * @param {Object} request - An object containing the following fields:
 *      <br>`name` : required - {string} The name of the new channel
 *      <br>`orderer` : required - {Orderer} object instance representing the
 *                      Orderer to send the create request
 *      <br>`envelope` : optional - byte[] of the envelope object containing all
 *                       required settings and signatures to initialize this channel.
 *                       This envelope would have been created by the command
 *                       line tool "configtx".
 *      <br>`config` : optional - {byte[]} Protobuf ConfigUpdate object extracted from
 *                     a ConfigEnvelope created by the ConfigTX tool.
 *                     see extractChannelConfig()
 *      <br>`signatures` : optional - {ConfigSignature[]} the list of collected signatures
 *                         required by the channel create policy when using the `config` parameter.
 * @returns {Result} Result Object with status on the create process.
 */
func (c *client) CreateChannel(request *api.CreateChannelRequest) error {
	haveEnvelope := false
	if request != nil && request.Envelope != nil {
		logger.Debug("createChannel - have envelope")
		haveEnvelope = true
	}
	return c.CreateOrUpdateChannel(request, haveEnvelope)
}

func (c *client) CreateOrUpdateChannel(request *api.CreateChannelRequest, haveEnvelope bool) error {
	// Validate request
	if request == nil {
		return fmt.Errorf("Missing all required input request parameters for initialize channel")
	}

	if request.Config == nil && !haveEnvelope {
		return fmt.Errorf("Missing envelope request parameter containing the configuration of the new channel")
	}

	if request.Signatures == nil && !haveEnvelope {
		return fmt.Errorf("Missing signatures request parameter for the new channel")
	}

	if request.TxID == "" && !haveEnvelope {
		return fmt.Errorf("Missing txId request parameter")
	}

	if request.Nonce == nil && !haveEnvelope {
		return fmt.Errorf("Missing nonce request parameter")
	}

	if request.Orderer == nil {
		return fmt.Errorf("Missing orderer request parameter for the initialize channel")
	}

	if request.Name == "" {
		return fmt.Errorf("Missing name request parameter for the new channel")
	}

	// channel = null;
	var signature []byte
	var payloadBytes []byte

	if haveEnvelope {
		logger.Debug("createOrUpdateChannel - have envelope")
		envelope := &common.Envelope{}
		err := proto.Unmarshal(request.Envelope, envelope)
		if err != nil {
			return fmt.Errorf("Error unmarshalling channel configuration data: %s", err.Error())
		}
		signature = envelope.Signature
		payloadBytes = envelope.Payload
	} else {
		logger.Debug("createOrUpdateChannel - have config_update")
		configUpdateEnvelope := &common.ConfigUpdateEnvelope{
			ConfigUpdate: request.Config,
			Signatures:   request.Signatures,
		}

		channelHeader, err := channel.BuildChannelHeader(common.HeaderType_CONFIG_UPDATE, request.Name, request.TxID, 0, "", time.Now())
		if err != nil {
			return fmt.Errorf("error when building channel header: %v", err)
		}
		creator, err := c.GetIdentity()
		if err != nil {
			return fmt.Errorf("Error getting creator: %v", err)
		}

		header, err := fc.BuildHeader(creator, channelHeader, request.Nonce)
		if err != nil {
			return fmt.Errorf("error when building header: %v", err)
		}
		configUpdateEnvelopeBytes, err := proto.Marshal(configUpdateEnvelope)
		if err != nil {
			return fmt.Errorf("error marshaling configUpdateEnvelope: %v", err)
		}
		payload := &common.Payload{
			Header: header,
			Data:   configUpdateEnvelopeBytes,
		}
		payloadBytes, err = proto.Marshal(payload)
		if err != nil {
			return fmt.Errorf("error marshaling payload: %v", err)
		}

		signature, err = fc.SignObjectWithKey(payloadBytes, c.userContext.GetPrivateKey(), &bccsp.SHAOpts{}, nil, c.GetCryptoSuite())
		if err != nil {
			return fmt.Errorf("error singing payload: %v", err)
		}
	}

	// Send request
	_, err := request.Orderer.SendBroadcast(&api.SignedEnvelope{
		Signature: signature,
		Payload:   payloadBytes,
	})
	if err != nil {
		return fmt.Errorf("Could not broadcast to orderer %s: %s", request.Orderer.GetURL(), err.Error())
	}

	return nil
}

//QueryChannels
/**
 * Queries the names of all the channels that a
 * peer has joined.
 * @param {Peer} peer
 * @returns {object} ChannelQueryResponse proto
 */
func (c *client) QueryChannels(peer api.Peer) (*pb.ChannelQueryResponse, error) {

	if peer == nil {
		return nil, fmt.Errorf("QueryChannels requires peer")
	}

	responses, err := channel.QueryByChaincode("cscc", []string{"GetChannels"}, []api.Peer{peer}, c)
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
func (c *client) QueryInstalledChaincodes(peer api.Peer) (*pb.ChaincodeQueryResponse, error) {

	if peer == nil {
		return nil, fmt.Errorf("To query installed chaincdes you need to pass peer")
	}
	responses, err := channel.QueryByChaincode("lscc", []string{"getinstalledchaincodes"}, []api.Peer{peer}, c)
	if err != nil {
		return nil, fmt.Errorf("Invoke lscc getinstalledchaincodes return error: %v", err)
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
	chaincodePackage []byte, targets []api.Peer) ([]*api.TransactionProposalResponse, string, error) {

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
		chaincodePackage, err = packager.PackageCC(chaincodePath, "")
		if err != nil {
			return nil, "", fmt.Errorf("PackageCC return error: %s", err)
		}
	}

	now := time.Now()
	cds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincodeName, Path: chaincodePath, Version: chaincodeVersion}},
		CodePackage: chaincodePackage, EffectiveDate: &google_protobuf.Timestamp{Seconds: int64(now.Second()), Nanos: int32(now.Nanosecond())}}

	creator, err := c.GetIdentity()
	if err != nil {
		return nil, "", fmt.Errorf("Error getting creator: %v", err)
	}

	// create an install from a chaincodeDeploymentSpec
	proposal, txID, err := protos_utils.CreateInstallProposalFromCDS(cds, creator)
	if err != nil {
		return nil, "", fmt.Errorf("Could not create chaincode Deploy proposal, err %s", err)
	}
	proposalBytes, err := protos_utils.GetBytesProposal(proposal)
	if err != nil {
		return nil, "", err
	}
	user, err := c.LoadUserFromStateStore("")
	if err != nil {
		return nil, "", fmt.Errorf("Error loading user from store: %s", err)
	}
	signature, err := fc.SignObjectWithKey(proposalBytes, user.GetPrivateKey(), &bccsp.SHAOpts{}, nil, c.GetCryptoSuite())
	if err != nil {
		return nil, "", err
	}

	signedProposal, err := &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}, nil
	if err != nil {
		return nil, "", err
	}

	transactionProposalResponse, err := channel.SendTransactionProposal(&api.TransactionProposal{
		SignedProposal: signedProposal,
		Proposal:       proposal,
		TransactionID:  txID,
	}, 0, targets)

	return transactionProposalResponse, txID, err
}

// GetIdentity returns client's serialized identity
func (c *client) GetIdentity() ([]byte, error) {

	if c.userContext == nil {
		return nil, fmt.Errorf("User is nil")
	}
	serializedIdentity := &msp.SerializedIdentity{Mspid: c.config.GetFabricCAID(),
		IdBytes: c.userContext.GetEnrollmentCertificate()}
	identity, err := proto.Marshal(serializedIdentity)
	if err != nil {
		return nil, fmt.Errorf("Could not Marshal serializedIdentity, err %s", err)
	}
	return identity, nil
}

// GetUserContext ...
func (c *client) GetUserContext() api.User {
	return c.userContext
}

// SetUserContext ...
func (c *client) SetUserContext(user api.User) {
	c.userContext = user
}

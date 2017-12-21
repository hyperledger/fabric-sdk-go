/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricclient

import (
	"encoding/json"
	"time"

	"github.com/golang/protobuf/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/timestamp"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/crypto"
	fcutils "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/util"
	protos_utils "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/utils"
	ccomm "github.com/hyperledger/fabric-sdk-go/pkg/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	channel "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	fc "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal/txnproc"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

// Client enables access to a Fabric network.
type Client struct {
	channels       map[string]fab.Channel
	cryptoSuite    apicryptosuite.CryptoSuite
	stateStore     fab.KeyValueStore
	userContext    fab.User
	config         config.Config
	signingManager fab.SigningManager
}

// NewClient returns a Client instance.
func NewClient(config config.Config) *Client {
	channels := make(map[string]fab.Channel)
	c := Client{channels: channels, cryptoSuite: nil, stateStore: nil, userContext: nil, config: config}
	return &c
}

// NewChannel returns a channel instance with the given name.
func (c *Client) NewChannel(name string) (fab.Channel, error) {
	if _, ok := c.channels[name]; ok {
		return nil, errors.Errorf("channel %s already exists", name)
	}
	var err error
	channel, err := channel.NewChannel(name, c)
	if err != nil {
		return nil, err
	}
	c.channels[name] = channel
	return c.channels[name], nil
}

// Config returns the configuration of the client.
func (c *Client) Config() config.Config {
	return c.config
}

// Channel returns the channel by ID
func (c *Client) Channel(id string) fab.Channel {
	return c.channels[id]
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
func (c *Client) QueryChannelInfo(name string, peers []fab.Peer) (fab.Channel, error) {
	return nil, errors.Errorf("Not implemented yet")
}

// SetStateStore ...
/*
 * The SDK should have a built-in key value store implementation (suggest a file-based implementation to allow easy setup during
 * development). But production systems would want a store backed by database for more robust storage and clustering,
 * so that multiple app instances can share app state via the database (note that this doesn’t necessarily make the app stateful).
 * This API makes this pluggable so that different store implementations can be selected by the application.
 */
func (c *Client) SetStateStore(stateStore fab.KeyValueStore) {
	c.stateStore = stateStore
}

// StateStore is a convenience method for obtaining the state store object in use for this client.
func (c *Client) StateStore() fab.KeyValueStore {
	return c.stateStore
}

// SetCryptoSuite is a convenience method for obtaining the state store object in use for this client.
func (c *Client) SetCryptoSuite(cryptoSuite apicryptosuite.CryptoSuite) {
	c.cryptoSuite = cryptoSuite
}

// CryptoSuite is a convenience method for obtaining the CryptoSuite object in use for this client.
func (c *Client) CryptoSuite() apicryptosuite.CryptoSuite {
	return c.cryptoSuite
}

// SigningManager returns the signing manager
func (c *Client) SigningManager() fab.SigningManager {
	return c.signingManager
}

// SetSigningManager is a convenience method to set signing manager
func (c *Client) SetSigningManager(signingMgr fab.SigningManager) {
	c.signingManager = signingMgr
}

// SaveUserToStateStore ...
/*
 * Sets an instance of the User class as the security context of this client instance. This user’s credentials (ECert) will be
 * used to conduct transactions and queries with the blockchain network. Upon setting the user context, the SDK saves the object
 * in a persistence cache if the “state store” has been set on the Client instance. If no state store has been set,
 * this cache will not be established and the application is responsible for setting the user context again when the application
 * crashed and is recovered.
 */
func (c *Client) SaveUserToStateStore(user fab.User, skipPersistence bool) error {
	if user == nil {
		return errors.New("user required")
	}

	if user.Name() == "" {
		return errors.New("user name is empty")
	}
	c.userContext = user
	if !skipPersistence {
		if c.stateStore == nil {
			return errors.New("stateStore is nil")
		}
		userJSON := &identity.JSON{
			MspID:                 user.MspID(),
			Roles:                 user.Roles(),
			PrivateKeySKI:         user.PrivateKey().SKI(),
			EnrollmentCertificate: user.EnrollmentCertificate(),
		}
		data, err := json.Marshal(userJSON)
		if err != nil {
			return errors.Wrap(err, "marshal json return error")
		}
		err = c.stateStore.SetValue(user.Name(), data)
		if err != nil {
			return errors.WithMessage(err, "stateStore SetValue failed")
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
func (c *Client) LoadUserFromStateStore(name string) (fab.User, error) {
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
		return nil, errors.New("cryptoSuite required")
	}
	value, err := c.stateStore.Value(name)
	if err != nil {
		return nil, nil
	}
	var userJSON identity.JSON
	err = json.Unmarshal(value, &userJSON)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal user JSON failed")
	}
	user := identity.NewUser(name, userJSON.MspID)
	user.SetRoles(userJSON.Roles)
	user.SetEnrollmentCertificate(userJSON.EnrollmentCertificate)
	key, err := c.cryptoSuite.GetKey(userJSON.PrivateKeySKI)
	if err != nil {
		return nil, errors.Wrap(err, "cryptoSuite GetKey failed")
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
func (c *Client) ExtractChannelConfig(configEnvelope []byte) ([]byte, error) {
	logger.Debug("extractConfigUpdate - start")

	envelope := &common.Envelope{}
	err := proto.Unmarshal(configEnvelope, envelope)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal config envelope failed")
	}

	payload := &common.Payload{}
	err = proto.Unmarshal(envelope.Payload, payload)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal envelope payload failed")
	}

	configUpdateEnvelope := &common.ConfigUpdateEnvelope{}
	err = proto.Unmarshal(payload.Data, configUpdateEnvelope)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal config update envelope")
	}

	return configUpdateEnvelope.ConfigUpdate, nil
}

// SignChannelConfig ...
/**
 * Sign a configuration
 * @param {byte[]} config - The Configuration Update in byte form
 * @return {ConfigSignature} - The signature of the current user on the config bytes
 */
func (c *Client) SignChannelConfig(config []byte, signer fab.User) (*common.ConfigSignature, error) {
	logger.Debug("SignChannelConfig - start")

	if config == nil {
		return nil, errors.New("channel configuration required")
	}

	signingUser := signer
	// If signing user is not provided default to client's user context
	if signingUser == nil {
		signingUser = c.userContext
	}

	if signingUser == nil {
		return nil, errors.New("user context required")
	}

	creator, err := signingUser.Identity()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get user context's identity")
	}
	nonce, err := fc.GenerateRandomNonce()
	if err != nil {
		return nil, errors.Wrap(err, "GenerateRandomNonce failed")
	}

	// signature is across a signature header and the config update
	signatureHeader := &common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
	}
	signatureHeaderBytes, err := proto.Marshal(signatureHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal signatureHeader failed")
	}

	signingMgr := c.SigningManager()
	if signingMgr == nil {
		return nil, errors.New("signing manager is nil")
	}

	// get all the bytes to be signed together, then sign
	signingBytes := fcutils.ConcatenateBytes(signatureHeaderBytes, config)
	signature, err := signingMgr.Sign(signingBytes, signingUser.PrivateKey())
	if err != nil {
		return nil, errors.WithMessage(err, "signing of channel config failed")
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
func (c *Client) CreateChannel(request fab.CreateChannelRequest) (apitxn.TransactionID, error) {
	haveEnvelope := false
	if request.Envelope != nil {
		logger.Debug("createChannel - have envelope")
		haveEnvelope = true
	}

	if !haveEnvelope && request.TxnID.ID == "" {
		txnID, err := c.NewTxnID()
		if err != nil {
			return txnID, err
		}
		request.TxnID = txnID
	}

	return request.TxnID, c.createOrUpdateChannel(request, haveEnvelope)
}

// createOrUpdateChannel creates a new channel or updates an existing channel.
func (c *Client) createOrUpdateChannel(request fab.CreateChannelRequest, haveEnvelope bool) error {
	// Validate request
	if request.Config == nil && !haveEnvelope {
		return errors.New("missing envelope request parameter containing the configuration of the new channel")
	}

	if request.Signatures == nil && !haveEnvelope {
		return errors.New("missing signatures request parameter for the new channel")
	}

	if request.TxnID.ID == "" && !haveEnvelope {
		return errors.New("txId required")
	}

	if request.TxnID.Nonce == nil && !haveEnvelope {
		return errors.New("nonce required")
	}

	if request.Orderer == nil {
		return errors.New("missing orderer request parameter for the initialize channel")
	}

	if request.Name == "" {
		return errors.New("missing name request parameter for the new channel")
	}

	// channel = null;
	var signature []byte
	var payloadBytes []byte

	if haveEnvelope {
		logger.Debug("createOrUpdateChannel - have envelope")
		envelope := &common.Envelope{}
		err := proto.Unmarshal(request.Envelope, envelope)
		if err != nil {
			return errors.Wrap(err, "unmarshal request envelope failed")
		}
		signature = envelope.Signature
		payloadBytes = envelope.Payload
	} else {
		logger.Debug("createOrUpdateChannel - have config_update")
		configUpdateEnvelope := &common.ConfigUpdateEnvelope{
			ConfigUpdate: request.Config,
			Signatures:   request.Signatures,
		}

		// TODO: Move
		tlsCertHash := ccomm.TLSCertHash(c.config)
		channelHeader, err := channel.BuildChannelHeader(common.HeaderType_CONFIG_UPDATE, request.Name, request.TxnID.ID, 0, "", time.Now(), tlsCertHash)
		if err != nil {
			return errors.WithMessage(err, "BuildChannelHeader failed")
		}
		if c.userContext == nil {
			return errors.New("user context is nil")
		}
		creator, err := c.userContext.Identity()
		if err != nil {
			return errors.WithMessage(err, "getting creator failed")
		}

		header, err := fc.BuildHeader(creator, channelHeader, request.TxnID.Nonce)
		if err != nil {
			return errors.Wrap(err, "BuildHeader failed")
		}
		configUpdateEnvelopeBytes, err := proto.Marshal(configUpdateEnvelope)
		if err != nil {
			return errors.Wrap(err, "marshal configUpdateEnvelope failed")
		}
		payload := &common.Payload{
			Header: header,
			Data:   configUpdateEnvelopeBytes,
		}
		payloadBytes, err = proto.Marshal(payload)
		if err != nil {
			return errors.Wrap(err, "marshal payload failed")
		}

		signingMgr := c.SigningManager()
		if signingMgr == nil {
			return errors.New("signing manager is nil")
		}

		signature, err = signingMgr.Sign(payloadBytes, c.UserContext().PrivateKey())
		if err != nil {
			return errors.WithMessage(err, "signing payload failed")
		}
	}

	// Send request
	_, err := request.Orderer.SendBroadcast(&fab.SignedEnvelope{
		Signature: signature,
		Payload:   payloadBytes,
	})
	if err != nil {
		return errors.WithMessage(err, "failed broadcast to orderer")
	}

	return nil
}

// QueryChannels queries the names of all the channels that a peer has joined.
func (c *Client) QueryChannels(peer fab.Peer) (*pb.ChannelQueryResponse, error) {

	if peer == nil {
		return nil, errors.New("peer required")
	}

	payload, err := c.queryBySystemChaincodeByTarget("cscc", "GetChannels", [][]byte{}, peer)
	if err != nil {
		return nil, errors.WithMessage(err, "cscc.GetChannels failed")
	}

	response := new(pb.ChannelQueryResponse)
	err = proto.Unmarshal(payload, response)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal ChannelQueryResponse failed")
	}
	return response, nil
}

// QueryInstalledChaincodes queries the installed chaincodes on a peer.
// Returns the details of all chaincodes installed on a peer.
func (c *Client) QueryInstalledChaincodes(peer fab.Peer) (*pb.ChaincodeQueryResponse, error) {

	if peer == nil {
		return nil, errors.New("peer required")
	}
	payload, err := c.queryBySystemChaincodeByTarget("lscc", "getinstalledchaincodes", [][]byte{}, peer)
	if err != nil {
		return nil, errors.WithMessage(err, "lscc.getinstalledchaincodes failed")
	}
	response := new(pb.ChaincodeQueryResponse)
	err = proto.Unmarshal(payload, response)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal ChaincodeQueryResponse failed")
	}

	return response, nil
}

// InstallChaincode sends an install proposal to one or more endorsing peers.
func (c *Client) InstallChaincode(req fab.InstallChaincodeRequest) ([]*apitxn.TransactionProposalResponse, string, error) {

	if req.Name == "" {
		return nil, "", errors.New("chaincode name required")
	}
	if req.Path == "" {
		return nil, "", errors.New("chaincode path required")
	}
	if req.Version == "" {
		return nil, "", errors.New("chaincode version required")
	}
	if req.Package == nil {
		return nil, "", errors.New("chaincode package is required")
	}

	now := time.Now()
	cds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: req.Package.Type, ChaincodeId: &pb.ChaincodeID{Name: req.Name, Path: req.Path, Version: req.Version}},
		CodePackage: req.Package.Code, EffectiveDate: &google_protobuf.Timestamp{Seconds: int64(now.Second()), Nanos: int32(now.Nanosecond())}}

	if c.userContext == nil {
		return nil, "", errors.New("user context required")
	}
	creator, err := c.userContext.Identity()
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to get creator identity")
	}

	// create an install from a chaincodeDeploymentSpec
	proposal, txID, err := protos_utils.CreateInstallProposalFromCDS(cds, creator)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to create chaincode deploy proposal")
	}
	proposalBytes, err := protos_utils.GetBytesProposal(proposal)
	if err != nil {
		return nil, "", err
	}
	user := c.UserContext()
	if user == nil {
		return nil, "", errors.New("User context is nil")
	}

	signingMgr := c.SigningManager()
	if signingMgr == nil {
		return nil, "", errors.Errorf("signing manager is nil")
	}

	signature, err := signingMgr.Sign(proposalBytes, user.PrivateKey())
	if err != nil {
		return nil, "", err
	}

	signedProposal := &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}

	txnID := apitxn.TransactionID{ID: txID} // Nonce is missing

	transactionProposalResponse, err := txnproc.SendTransactionProposalToProcessors(&apitxn.TransactionProposal{
		SignedProposal: signedProposal,
		Proposal:       proposal,
		TxnID:          txnID,
	}, req.Targets)

	return transactionProposalResponse, txID, err
}

// UserContext returns the current User.
func (c *Client) UserContext() fab.User {
	return c.userContext
}

// SetUserContext ...
func (c *Client) SetUserContext(user fab.User) {
	c.userContext = user
}

// NewTxnID computes a TransactionID for the current user context
func (c *Client) NewTxnID() (apitxn.TransactionID, error) {
	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return apitxn.TransactionID{}, err
	}

	if c.userContext == nil {
		return apitxn.TransactionID{}, errors.New("user context is nil")
	}
	creator, err := c.userContext.Identity()
	if err != nil {
		return apitxn.TransactionID{}, err
	}

	id, err := protos_utils.ComputeProposalTxID(nonce, creator)
	if err != nil {
		return apitxn.TransactionID{}, err
	}

	txnID := apitxn.TransactionID{
		ID:    id,
		Nonce: nonce,
	}

	return txnID, nil
}

func (c *Client) queryBySystemChaincodeByTarget(chaincodeID string, fcn string, args [][]byte, target apitxn.ProposalProcessor) ([]byte, error) {
	targets := []apitxn.ProposalProcessor{target}
	request := apitxn.ChaincodeInvokeRequest{
		ChaincodeID: chaincodeID,
		Fcn:         fcn,
		Args:        args,
		Targets:     targets,
	}
	responses, err := channel.QueryBySystemChaincode(request, c)

	if err != nil {
		return nil, errors.WithMessage(err, "QueryBySystemChaincode failed")
	}
	// we are only querying one peer hence one result
	if len(responses) != 1 {
		return nil, errors.Errorf("QueryBySystemChaincode should have one result only - actual result count: %d", len(responses))
	}

	return responses[0], nil

}

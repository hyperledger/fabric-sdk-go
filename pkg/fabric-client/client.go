/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricclient

import (
	"encoding/json"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/kvstore"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	channel "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// Client enables access to a Fabric network.
type Client struct {
	channels        map[string]fab.Channel
	cryptoSuite     apicryptosuite.CryptoSuite
	stateStore      kvstore.KVStore
	signingIdentity fab.IdentityContext
	config          config.Config
	signingManager  fab.SigningManager
}

type fabContext struct {
	fab.ProviderContext
	fab.IdentityContext
}

// NewClient returns a Client instance.
//
// Deprecated: see fabsdk package.
func NewClient(config config.Config) *Client {
	channels := make(map[string]fab.Channel)
	c := Client{channels: channels, config: config}
	return &c
}

// NewChannel returns a channel instance with the given name.
//
// Deprecated: see fabsdk package.
func (c *Client) NewChannel(name string) (fab.Channel, error) {
	if _, ok := c.channels[name]; ok {
		return nil, errors.Errorf("channel %s already exists", name)
	}

	ctx := fabContext{ProviderContext: c, IdentityContext: c.signingIdentity}
	channel, err := channel.New(ctx, chconfig.NewChannelCfg(name))
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
//
// Deprecated: see fabsdk package.
/*
 * The SDK should have a built-in key value store implementation (suggest a file-based implementation to allow easy setup during
 * development). But production systems would want a store backed by database for more robust kvstore and clustering,
 * so that multiple app instances can share app state via the database (note that this doesn’t necessarily make the app stateful).
 * This API makes this pluggable so that different store implementations can be selected by the application.
 */
func (c *Client) SetStateStore(stateStore kvstore.KVStore) {
	c.stateStore = stateStore
}

// StateStore is a convenience method for obtaining the state store object in use for this client.
func (c *Client) StateStore() kvstore.KVStore {
	return c.stateStore
}

// SetCryptoSuite is a convenience method for obtaining the state store object in use for this client.
//
// Deprecated: see fabsdk package.
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
//
// Deprecated: see fabsdk package.
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
func (c *Client) SaveUserToStateStore(user fab.User) error {
	if user == nil {
		return errors.New("user required")
	}

	if user.Name() == "" {
		return errors.New("user name is empty")
	}

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
	err = c.stateStore.Store(user.Name(), data)
	if err != nil {
		return errors.WithMessage(err, "stateStore SetValue failed")
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
	if name == "" {
		return nil, nil
	}
	if c.stateStore == nil {
		return nil, nil
	}
	if c.cryptoSuite == nil {
		return nil, errors.New("cryptoSuite required")
	}
	value, err := c.stateStore.Load(name)
	if err != nil {
		if err == kvstore.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	valueBytes, ok := value.([]byte)
	if !ok {
		return nil, errors.New("state store return wrong data type")
	}
	var userJSON identity.JSON
	err = json.Unmarshal(valueBytes, &userJSON)
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
	return user, nil
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
	ctx := fabContext{ProviderContext: c, IdentityContext: c.signingIdentity}
	rc := resource.New(ctx)
	return rc.ExtractChannelConfig(configEnvelope)
}

// SignChannelConfig ...
/**
 * Sign a configuration
 * @param {byte[]} config - The Configuration Update in byte form
 * @return {ConfigSignature} - The signature of the current user on the config bytes
 */
func (c *Client) SignChannelConfig(config []byte, signer fab.IdentityContext) (*common.ConfigSignature, error) {
	ctx := fabContext{ProviderContext: c, IdentityContext: c.signingIdentity}
	rc := resource.New(ctx)
	return rc.SignChannelConfig(config, signer)
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
func (c *Client) CreateChannel(request fab.CreateChannelRequest) (fab.TransactionID, error) {
	ctx := fabContext{ProviderContext: c, IdentityContext: c.signingIdentity}
	rc := resource.New(ctx)
	return rc.CreateChannel(request)
}

// QueryChannels queries the names of all the channels that a peer has joined.
func (c *Client) QueryChannels(peer fab.Peer) (*pb.ChannelQueryResponse, error) {
	ctx := fabContext{ProviderContext: c, IdentityContext: c.signingIdentity}
	rc := resource.New(ctx)
	return rc.QueryChannels(peer)
}

// QueryInstalledChaincodes queries the installed chaincodes on a peer.
// Returns the details of all chaincodes installed on a peer.
func (c *Client) QueryInstalledChaincodes(peer fab.Peer) (*pb.ChaincodeQueryResponse, error) {
	ctx := fabContext{ProviderContext: c, IdentityContext: c.signingIdentity}
	rc := resource.New(ctx)
	return rc.QueryInstalledChaincodes(peer)
}

// InstallChaincode sends an install proposal to one or more endorsing peers.
func (c *Client) InstallChaincode(req fab.InstallChaincodeRequest) ([]*fab.TransactionProposalResponse, string, error) {
	ctx := fabContext{ProviderContext: c, IdentityContext: c.signingIdentity}
	rc := resource.New(ctx)
	return rc.InstallChaincode(req)
}

// IdentityContext returns the current identity for signing.
func (c *Client) IdentityContext() fab.IdentityContext {
	return c.signingIdentity
}

// SetIdentityContext sets the identity for signing
func (c *Client) SetIdentityContext(user fab.IdentityContext) {
	c.signingIdentity = user
}

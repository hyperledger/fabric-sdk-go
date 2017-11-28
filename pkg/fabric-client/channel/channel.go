/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

// Channel  captures settings for a channel, which is created by
// the orderers to isolate transactions delivery to peers participating on channel.
type Channel struct {
	name          string // aka channel ID
	peers         map[string]fab.Peer
	orderers      map[string]fab.Orderer
	clientContext ClientContext
	primaryPeer   fab.Peer
	mspManager    msp.MSPManager
	anchorPeers   []*fab.OrgAnchorPeer
	initialized   bool
}

// ClientContext ...
type ClientContext interface {
	UserContext() fab.User
	SigningManager() fab.SigningManager
	NewTxnID() (apitxn.TransactionID, error)
	Config() config.Config
	CryptoSuite() apicryptosuite.CryptoSuite
}

// NewChannel represents a channel in a Fabric network.
// name: used to identify different channel instances. The naming of channel instances
// is enforced by the ordering service and must be unique within the blockchain network.
// client: Provides operational context such as submitting User etc.
func NewChannel(name string, client fab.FabricClient) (*Channel, error) {
	if name == "" {
		return nil, errors.Errorf("name is required")
	}
	if client == nil {
		return nil, errors.Errorf("client is required")
	}
	p := make(map[string]fab.Peer)
	o := make(map[string]fab.Orderer)
	c := Channel{name: name, peers: p,
		orderers: o, clientContext: client, mspManager: msp.NewMSPManager()}
	logger.Infof("Constructed channel instance for channel %s", c.name)
	logger.Debugf("Constructed channel instance: %V", c)

	return &c, nil
}

// ClientContext returns the Client that was passed in to NewChannel
func (c *Channel) ClientContext() ClientContext {
	return c.clientContext
}

// Name returns the channel name.
func (c *Channel) Name() string {
	return c.name
}

// AddPeer adds a peer endpoint to channel.
// It returns error if the peer with that url already exists.
func (c *Channel) AddPeer(peer fab.Peer) error {
	url := peer.URL()
	if c.peers[url] != nil {
		return errors.Errorf("peer with URL %s already exists", url)
	}
	c.peers[url] = peer
	return nil
}

// RemovePeer remove a peer endpoint from channel.
func (c *Channel) RemovePeer(peer fab.Peer) {
	url := peer.URL()
	if c.peers[url] != nil {
		delete(c.peers, url)
		logger.Debugf("Removed peer with URL %s", url)
	}
}

// Peers returns the peers of of the channel.
func (c *Channel) Peers() []fab.Peer {
	var peersArray []fab.Peer
	for _, v := range c.peers {
		peersArray = append(peersArray, v)
	}
	return peersArray
}

// AnchorPeers returns the anchor peers for this channel.
// Note: channel.Initialize() must be called first to retrieve anchor peers
func (c *Channel) AnchorPeers() []fab.OrgAnchorPeer {
	anchors := []fab.OrgAnchorPeer{}
	for _, anchor := range c.anchorPeers {
		anchors = append(anchors, *anchor)
	}

	return anchors
}

// SetPrimaryPeer sets the primary peer -- The peer to use for doing queries.
// Peer must be a peer on this channel's peer list.
// Default: When no primary peer has been set the first peer
// on the list will be used.
// It returns error when peer is not on the existing peer list
func (c *Channel) SetPrimaryPeer(peer fab.Peer) error {

	if !c.isValidPeer(peer) {
		return errors.New("primary peer must be on this channel peer list")
	}

	c.primaryPeer = c.peers[peer.URL()]
	return nil
}

// PrimaryPeer gets the primary peer -- the peer to use for doing queries.
// Default: When no primary peer has been set the first peer
// from map range will be used.
func (c *Channel) PrimaryPeer() fab.Peer {

	if c.primaryPeer != nil {
		return c.primaryPeer
	}

	// When no primary peer has been set default to the first peer
	// from map range - order is not guaranteed
	for _, peer := range c.peers {
		logger.Debugf("Primary peer was not set, using %s", peer.URL())
		return peer
	}

	return nil
}

// AddOrderer adds an orderer endpoint to a channel object, this is a local-only operation.
// A channel instance may choose to use a single orderer node, which will broadcast
// requests to the rest of the orderer network. Or if the application does not trust
// the orderer nodes, it can choose to use more than one by adding them to the channel instance.
// All APIs concerning the orderer will broadcast to all orderers simultaneously.
// orderer: An instance of the Orderer interface.
// Returns error if the orderer with that url already exists.
func (c *Channel) AddOrderer(orderer fab.Orderer) error {
	url := orderer.URL()
	if c.orderers[url] != nil {
		return errors.Errorf("orderer with URL %s already exists", url)
	}
	c.orderers[orderer.URL()] = orderer
	return nil
}

// RemoveOrderer removes orderer endpoint from a channel object, this is a local-only operation.
// orderer: An instance of the Orderer class.
func (c *Channel) RemoveOrderer(orderer fab.Orderer) {
	url := orderer.URL()
	if c.orderers[url] != nil {
		delete(c.orderers, url)
		logger.Debugf("Removed orderer with URL %s", url)
	}
}

// Orderers gets the orderers of a channel.
func (c *Channel) Orderers() []fab.Orderer {
	var orderersArray []fab.Orderer
	for _, v := range c.orderers {
		orderersArray = append(orderersArray, v)
	}
	return orderersArray
}

// SetMSPManager sets the MSP Manager for this channel.
// This utility method will not normally be used as the
// "Initialize()" method will read this channel's
// current configuration and reset the MSPManager with
// the MSP's found.
func (c *Channel) SetMSPManager(mspManager msp.MSPManager) {
	c.mspManager = mspManager
}

// MSPManager returns the MSP Manager for this channel
func (c *Channel) MSPManager() msp.MSPManager {
	return c.mspManager
}

// OrganizationUnits - to get identifier for the organization configured on the channel
func (c *Channel) OrganizationUnits() ([]string, error) {
	channelMSPManager := c.MSPManager()
	msps, err := channelMSPManager.GetMSPs()
	if err != nil {
		logger.Info("Cannot get channel manager")
		return nil, errors.WithMessage(err, "organization units were not set")
	}
	var orgIdentifiers []string
	for _, v := range msps {
		orgName, err := v.GetIdentifier()
		if err != nil {
			logger.Info("Organization does not have an identifier")
		}
		orgIdentifiers = append(orgIdentifiers, orgName)
	}
	return orgIdentifiers, nil
}

// Utility function to ensure that a peer exists on this channel.
// It returns true if peer exists on this channel
func (c *Channel) isValidPeer(peer fab.Peer) bool {
	return peer != nil && c.peers[peer.URL()] != nil
}

// TODO
// The following functions haven't been implemented.

// UpdateChannel calls the orderer(s) to update an existing channel. This allows the addition and
// deletion of Peer nodes to an existing channel, as well as the update of Peer
// certificate information upon certificate renewals.
// It returns whether or not the channel update process was successful.
func (c *Channel) UpdateChannel() bool {
	return false
}

// IsReadonly gets channel status to see if the underlying channel has been terminated,
// making it a read-only channel, where information (transactions and states)
// can be queried but no new transactions can be submitted.
// It returns read-only, true or not.
func (c *Channel) IsReadonly() bool {
	return false //to do
}

// IsInitialized ... TODO
func (c *Channel) IsInitialized() bool {
	return c.initialized
}

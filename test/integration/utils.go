/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"math/rand"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
)

// GenerateRandomID generates random ID
func GenerateRandomID() string {
	rand.Seed(time.Now().UnixNano())
	return randomString(10)
}

// Utility to create random string of strlen length
func randomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// InitializeChannel ...
func InitializeChannel(sdk *fabsdk.FabricSDK, orgID string, req chmgmt.SaveChannelRequest, targets []fab.ProposalProcessor) error {
	joinedTargets, err := FilterTargetsJoinedChannel(sdk, orgID, req.ChannelID, targets)
	if err != nil {
		return errors.WithMessage(err, "checking for joined targets failed")
	}

	if len(joinedTargets) != len(targets) {
		_, err := CreateChannel(sdk, req)
		if err != nil {
			return errors.Wrapf(err, "create channel failed")
		}

		_, err = JoinChannel(sdk, req.ChannelID)
		if err != nil {
			return errors.Wrapf(err, "join channel failed")
		}
	}
	return nil
}

// FilterTargetsJoinedChannel filters targets to those that have joined the named channel.
func FilterTargetsJoinedChannel(sdk *fabsdk.FabricSDK, orgID string, channelID string, targets []fab.ProposalProcessor) ([]fab.ProposalProcessor, error) {
	joinedTargets := []fab.ProposalProcessor{}
	session, err := sdk.NewClient(fabsdk.WithUser("Admin"), fabsdk.WithOrg(orgID)).Session()
	if err != nil {
		return nil, errors.WithMessage(err, "failed getting admin user session for org")
	}

	sc, err := sdk.FabricProvider().CreateResourceClient(session)
	if err != nil {
		return nil, errors.WithMessage(err, "NewResourceClient failed")
	}

	for _, target := range targets {
		// Check if primary peer has joined channel
		alreadyJoined, err := HasPeerJoinedChannel(sc, target, channelID)
		if err != nil {
			return nil, errors.WithMessage(err, "failed while checking if primary peer has already joined channel")
		}
		if alreadyJoined {
			joinedTargets = append(joinedTargets, target)
		}
	}
	return joinedTargets, nil
}

// CreateChannel attempts to save the named channel.
func CreateChannel(sdk *fabsdk.FabricSDK, req chmgmt.SaveChannelRequest) (bool, error) {

	// Channel management client is responsible for managing channels (create/update)
	chMgmtClient, err := sdk.NewClient(fabsdk.WithUser("Admin"), fabsdk.WithOrg("ordererorg")).ChannelMgmt()
	if err != nil {
		return false, errors.WithMessage(err, "Failed to create new channel management client")
	}

	// Create channel (or update if it already exists)
	if err = chMgmtClient.SaveChannel(req); err != nil {
		return false, nil
	}

	time.Sleep(time.Second * 5)
	return true, nil
}

// JoinChannel attempts to save the named channel.
func JoinChannel(sdk *fabsdk.FabricSDK, name string) (bool, error) {
	// Resource management client is responsible for managing resources (joining channels, install/instantiate/upgrade chaincodes)
	resMgmtClient, err := sdk.NewClient(fabsdk.WithUser("Admin")).ResourceMgmt()
	if err != nil {
		return false, errors.WithMessage(err, "Failed to create new resource management client")
	}

	if err = resMgmtClient.JoinChannel(name); err != nil {
		return false, nil
	}
	return true, nil
}

// CreateProposalProcessors initializes target peers based on config
func CreateProposalProcessors(config apiconfig.Config, orgs []string) ([]fab.ProposalProcessor, error) {
	peers := []fab.ProposalProcessor{}
	for _, org := range orgs {
		peerConfig, err := config.PeersConfig(org)
		if err != nil {
			return nil, errors.WithMessage(err, "reading peer config failed")
		}
		for _, p := range peerConfig {
			endorser, err := peer.New(config, peer.FromPeerConfig(&apiconfig.NetworkPeer{PeerConfig: p}))
			if err != nil {
				return nil, errors.WithMessage(err, "NewPeer failed")
			}
			peers = append(peers, endorser)
			if err != nil {
				return nil, errors.WithMessage(err, "adding peer failed")
			}
		}
	}
	return peers, nil
}

// HasPeerJoinedChannel checks whether the peer has already joined the channel.
// It returns true if it has, false otherwise, or an error
func HasPeerJoinedChannel(client fab.Resource, peer fab.ProposalProcessor, channel string) (bool, error) {
	foundChannel := false
	response, err := client.QueryChannels(peer)
	if err != nil {
		return false, errors.WithMessage(err, "failed to query channel for peer")
	}
	for _, responseChannel := range response.Channels {
		if responseChannel.ChannelId == channel {
			foundChannel = true
		}
	}

	return foundChannel, nil
}

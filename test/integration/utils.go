/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
)

const (
	adminUser      = "Admin"
	ordererOrgName = "ordererorg"
)

// GenerateRandomID generates random ID
func GenerateRandomID() string {
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
func InitializeChannel(sdk *fabsdk.FabricSDK, orgID string, req resmgmt.SaveChannelRequest, targets []string) error {

	joinedTargets, err := FilterTargetsJoinedChannel(sdk, orgID, req.ChannelID, targets)
	if err != nil {
		return errors.WithMessage(err, "checking for joined targets failed")
	}

	if len(joinedTargets) != len(targets) {
		_, err := CreateChannel(sdk, req)
		if err != nil {
			return errors.Wrapf(err, "create channel failed")
		}

		_, err = JoinChannel(sdk, req.ChannelID, orgID)
		if err != nil {
			return errors.Wrapf(err, "join channel failed")
		}
	}
	return nil
}

// FilterTargetsJoinedChannel filters targets to those that have joined the named channel.
func FilterTargetsJoinedChannel(sdk *fabsdk.FabricSDK, orgID string, channelID string, targets []string) ([]string, error) {
	var joinedTargets []string

	//prepare context
	clientContext := sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(orgID))

	rc, err := resmgmt.New(clientContext)
	if err != nil {
		return nil, errors.WithMessage(err, "failed getting admin user session for org")
	}

	for _, target := range targets {
		// Check if primary peer has joined channel
		alreadyJoined, err := HasPeerJoinedChannel(rc, target, channelID)
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
func CreateChannel(sdk *fabsdk.FabricSDK, req resmgmt.SaveChannelRequest) (bool, error) {

	//prepare context
	clientContext := sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(ordererOrgName))

	// Channel management client is responsible for managing channels (create/update)
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return false, errors.WithMessage(err, "Failed to create new channel management client")
	}

	// Create channel (or update if it already exists)
	if err = resMgmtClient.SaveChannel(req); err != nil {
		return false, err
	}

	time.Sleep(time.Second * 5)
	return true, nil
}

// JoinChannel attempts to save the named channel.
func JoinChannel(sdk *fabsdk.FabricSDK, name, orgID string) (bool, error) {
	//prepare context
	clientContext := sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(orgID))

	// Resource management client is responsible for managing resources (joining channels, install/instantiate/upgrade chaincodes)
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return false, errors.WithMessage(err, "Failed to create new resource management client")
	}

	if err = resMgmtClient.JoinChannel(name); err != nil {
		return false, nil
	}
	return true, nil
}

// OrgTargetPeers determines peer endpoints for orgs
func OrgTargetPeers(config core.Config, orgs []string) ([]string, error) {
	var peers []string
	for _, org := range orgs {
		peerConfig, err := config.PeersConfig(org)
		if err != nil {
			return nil, errors.WithMessage(err, "reading peer config failed")
		}
		for _, p := range peerConfig {
			peers = append(peers, p.URL)
		}
	}
	return peers, nil
}

// HasPeerJoinedChannel checks whether the peer has already joined the channel.
// It returns true if it has, false otherwise, or an error
func HasPeerJoinedChannel(client *resmgmt.Client, target string, channel string) (bool, error) {
	foundChannel := false
	response, err := client.QueryChannels(resmgmt.WithTargetURLs(target))
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

// CleanupTestPath removes the contents of a state store.
func CleanupTestPath(t *testing.T, storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		if t == nil {
			panic(fmt.Sprintf("Cleaning up directory '%s' failed: %v", storePath, err))
		}
		t.Fatalf("Cleaning up directory '%s' failed: %v", storePath, err)
	}
}

// CleanupUserData removes user data.
func CleanupUserData(t *testing.T, sdk *fabsdk.FabricSDK) {
	netConfig := sdk.Config()
	keyStorePath := netConfig.KeyStorePath()
	credentialStorePath := netConfig.CredentialStorePath()
	CleanupTestPath(t, keyStorePath)
	CleanupTestPath(t, credentialStorePath)
}

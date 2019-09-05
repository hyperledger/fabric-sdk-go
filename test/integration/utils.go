/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	fabApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/pkg/errors"
)

const (
	adminUser       = "Admin"
	ordererOrgName  = "OrdererOrg"
	ordererEndpoint = "orderer.example.com"
)

// GenerateRandomID generates random ID
func GenerateRandomID() string {
	return randomString(10)
}

// Utility to create random string of strlen length
func randomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	seed := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(seed)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rnd.Intn(len(chars))]
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
		_, err := SaveChannel(sdk, req)
		if err != nil {
			return errors.Wrapf(err, "create channel failed")
		}

		_, err = JoinChannel(sdk, req.ChannelID, orgID, targets)
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

// SaveChannel attempts to save (create or update) the named channel.
func SaveChannel(sdk *fabsdk.FabricSDK, req resmgmt.SaveChannelRequest) (bool, error) {

	//prepare context
	clientContext := sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(ordererOrgName))

	// Channel management client is responsible for managing channels (create/update)
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return false, errors.WithMessage(err, "Failed to create new channel management client")
	}

	// Create channel (or update if it already exists)
	if _, err = resMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(ordererEndpoint)); err != nil {
		return false, err
	}

	return true, nil
}

// JoinChannel attempts to save the named channel.
func JoinChannel(sdk *fabsdk.FabricSDK, name, orgID string, targets []string) (bool, error) {
	//prepare context
	clientContext := sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(orgID))

	// Resource management client is responsible for managing resources (joining channels, install/instantiate/upgrade chaincodes)
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return false, errors.WithMessage(err, "Failed to create new resource management client")
	}

	if err := resMgmtClient.JoinChannel(
		name,
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithTargetEndpoints(targets...),
		resmgmt.WithOrdererEndpoint(ordererEndpoint)); err != nil {
		return false, nil
	}
	return true, nil
}

// OrgTargetPeers determines peer endpoints for orgs
func OrgTargetPeers(orgs []string, configBackend ...core.ConfigBackend) ([]string, error) {
	networkConfig := fabApi.NetworkConfig{}
	err := lookup.New(configBackend...).UnmarshalKey("organizations", &networkConfig.Organizations)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get organizations from config ")
	}

	var peers []string
	for _, org := range orgs {
		orgConfig, ok := networkConfig.Organizations[strings.ToLower(org)]
		if !ok {
			continue
		}
		peers = append(peers, orgConfig.Peers...)
	}
	return peers, nil
}

// SetupMultiOrgContext creates an OrgContext for two organizations in the org channel.
func SetupMultiOrgContext(sdk *fabsdk.FabricSDK, org1Name string, org2Name string, org1AdminUser string, org2AdminUser string) ([]*OrgContext, error) {
	org1AdminContext := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1Name))
	org1ResMgmt, err := resmgmt.New(org1AdminContext)
	if err != nil {
		return nil, err
	}

	org1MspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(org1Name))
	if err != nil {
		return nil, err
	}
	org1Admin, err := org1MspClient.GetSigningIdentity(org1AdminUser)
	if err != nil {
		return nil, err
	}

	org2AdminContext := sdk.Context(fabsdk.WithUser(org2AdminUser), fabsdk.WithOrg(org2Name))
	org2ResMgmt, err := resmgmt.New(org2AdminContext)
	if err != nil {
		return nil, err
	}

	org2MspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(org2Name))
	if err != nil {
		return nil, err
	}
	org2Admin, err := org2MspClient.GetSigningIdentity(org2AdminUser)
	if err != nil {
		return nil, err
	}

	// Ensure that Gossip has propagated its view of local peers before invoking
	// install since some peers may be missed if we call InstallCC too early
	org1Peers, err := DiscoverLocalPeers(org1AdminContext, 2)
	if err != nil {
		return nil, errors.WithMessage(err, "discovery of local peers failed")
	}
	org2Peers, err := DiscoverLocalPeers(org2AdminContext, 2)
	if err != nil {
		return nil, errors.WithMessage(err, "discovery of local peers failed")
	}

	return []*OrgContext{
		{
			OrgID:                org1Name,
			CtxProvider:          org1AdminContext,
			ResMgmt:              org1ResMgmt,
			Peers:                org1Peers,
			SigningIdentity:      org1Admin,
			AnchorPeerConfigFile: "orgchannelOrg1MSPanchors.tx",
		},
		{
			OrgID:                org2Name,
			CtxProvider:          org2AdminContext,
			ResMgmt:              org2ResMgmt,
			Peers:                org2Peers,
			SigningIdentity:      org2Admin,
			AnchorPeerConfigFile: "orgchannelOrg2MSPanchors.tx",
		},
	}, nil
}

// HasPeerJoinedChannel checks whether the peer has already joined the channel.
// It returns true if it has, false otherwise, or an error
func HasPeerJoinedChannel(client *resmgmt.Client, target string, channel string) (bool, error) {
	foundChannel := false
	response, err := client.QueryChannels(resmgmt.WithTargetEndpoints(target), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
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
	var keyStorePath, credentialStorePath string

	configBackend, err := sdk.Config()
	if err != nil {
		// if an error is returned from Config, it means configBackend was nil, in this case simply hard code
		// the keyStorePath and credentialStorePath to the default values
		// This case is mostly happening due to configless test that is not passing a ConfigProvider to the SDK
		// which makes configBackend = nil.
		// Since configless test uses the same config values as the default ones (config_test.yaml), it's safe to
		// hard code these paths here
		keyStorePath = "/tmp/msp/keystore"
		credentialStorePath = "/tmp/state-store"
	} else {
		cryptoSuiteConfig := cryptosuite.ConfigFromBackend(configBackend)
		identityConfig, err := msp.ConfigFromBackend(configBackend)
		if err != nil {
			t.Fatal(err)
		}

		keyStorePath = cryptoSuiteConfig.KeyStorePath()
		credentialStorePath = identityConfig.CredentialStorePath()
	}

	CleanupTestPath(t, keyStorePath)
	CleanupTestPath(t, credentialStorePath)
}

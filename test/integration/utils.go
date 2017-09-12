/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"math/rand"
	"time"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
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

// HasPrimaryPeerJoinedChannel checks whether the primary peer of a channel
// has already joined the channel. It returns true if it has, false otherwise,
// or an error
func HasPrimaryPeerJoinedChannel(client fab.FabricClient, channel fab.Channel) (bool, error) {
	foundChannel := false
	primaryPeer := channel.PrimaryPeer()
	response, err := client.QueryChannels(primaryPeer)
	if err != nil {
		return false, fmt.Errorf("Error querying channel for primary peer: %s", err)
	}
	for _, responseChannel := range response.Channels {
		if responseChannel.ChannelId == channel.Name() {
			foundChannel = true
		}
	}

	return foundChannel, nil
}

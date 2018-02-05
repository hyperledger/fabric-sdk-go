/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"math/rand"
	"time"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
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

// HasPeerJoinedChannel checks whether the peer has already joined the channel.
// It returns true if it has, false otherwise, or an error
func HasPeerJoinedChannel(client fab.Resource, peer fab.Peer, channel string) (bool, error) {
	foundChannel := false
	response, err := client.QueryChannels(peer)
	if err != nil {
		return false, errors.WithMessage(err, "failed to query channel for primary peer")
	}
	for _, responseChannel := range response.Channels {
		if responseChannel.ChannelId == channel {
			foundChannel = true
		}
	}

	return foundChannel, nil
}

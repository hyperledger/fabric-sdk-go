/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package greylist

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/stretchr/testify/assert"
)

func TestGreylistFilter(t *testing.T) {
	expiryPeriod := time.Second * 4
	goodPeers := createMockPeers(0, 100)
	badPeers := createMockPeers(100, 200)

	f := New(expiryPeriod)
	for index, badPeer := range badPeers {
		f.Greylist(connectionFailedStatus(badPeer.URL()))
		assert.False(t, f.Accept(badPeer), "Expected bad peer to be greylisted")
		assert.True(t, f.Accept(goodPeers[index]), "Expected good peer to be accepted")
	}

	time.Sleep(expiryPeriod)
	for index, badPeer := range badPeers {
		assert.True(t, f.Accept(badPeer), "Expected bad peer to be accepted after expiry period")
		assert.True(t, f.Accept(goodPeers[index]), "Expected good peer to be accepted")
	}
}

func TestGreylistInvalidErr(t *testing.T) {
	f := New(time.Microsecond * 1)
	f.Greylist(fmt.Errorf("test"))

	ok, url := required(status.New(status.UnknownStatus, status.OK.ToInt32(), "", nil))
	assert.False(t, ok)
	assert.Empty(t, url)

	ok, url = required(status.New(status.EndorserClientStatus, status.ConnectionFailed.ToInt32(), "", nil))
	assert.True(t, ok)
	assert.Empty(t, url)
}

func connectionFailedStatus(url string) error {
	return status.New(status.EndorserClientStatus, status.ConnectionFailed.ToInt32(),
		"test", []interface{}{url})
}

func createMockPeers(fromIndex int, toIndex int) []fab.Peer {
	var mockPeers []fab.Peer
	for i := fromIndex; i < toIndex; i++ {
		mockPeers = append(mockPeers, mocks.NewMockPeer(fmt.Sprintf("testPeer%d", i),
			"grpcs://myPeer.org:"+strconv.Itoa(i)))
	}
	return mockPeers
}

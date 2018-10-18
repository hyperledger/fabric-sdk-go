/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package balancedsorter

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/balancer"
	fab "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/stretchr/testify/assert"
)

const (
	peer1URL = "peer1.org1.com:9999"
	peer2URL = "peer2.org1.com:9999"
	peer3URL = "peer3.org1.com:9999"
	peer4URL = "peer4.org1.com:9999"
	peer5URL = "peer5.org1.com:9999"
	peer6URL = "peer6.org1.com:9999"
)

var (
	peer1 = mocks.NewMockPeer("p1", peer1URL)
	peer2 = mocks.NewMockPeer("p2", peer2URL)
	peer3 = mocks.NewMockPeer("p3", peer3URL)
	peer4 = mocks.NewMockPeer("p4", peer4URL)
	peer5 = mocks.NewMockPeer("p5", peer5URL)
	peer6 = mocks.NewMockPeer("p6", peer6URL)

	allPeers = []fab.Peer{peer1, peer2, peer3, peer4, peer5, peer6}
)

func TestBalancedSorterDistributionRoundRobin(t *testing.T) {
	sort := New(WithBalancer(balancer.RoundRobin()))

	iterations := 100
	// Each peer should be called exactly the same number of times
	expectedMin := iterations / len(allPeers)

	count := make(map[string]int)

	for i := 0; i < iterations; i++ {
		peers := sort(allPeers)
		p := peers[0]
		count[p.URL()] = count[p.URL()] + 1
	}

	for url, c := range count {
		assert.Truef(t, c >= expectedMin, "Expecting peer [%s] to have been called at least %d times but got only %d", url, expectedMin, c)
	}
}

func TestBalancedSorterDistributionRandom(t *testing.T) {
	sort := New(WithBalancer(balancer.Random()))

	iterations := 1000
	expectedMin := iterations/len(allPeers) - 50

	count := make(map[string]int)

	for i := 0; i < iterations; i++ {
		peers := sort(allPeers)
		p := peers[0]
		count[p.URL()] = count[p.URL()] + 1
	}

	for url, c := range count {
		assert.Truef(t, c >= expectedMin, "Expecting peer [%s] to have been called at least %d times but got only %d", url, expectedMin, c)
	}
}

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())
	os.Exit(m.Run())
}

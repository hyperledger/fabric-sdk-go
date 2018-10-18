/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockheightsorter

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/balancer"
	fab "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	emocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
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
	peer1 = emocks.NewMockPeer("p1", peer1URL, 1000)
	peer2 = emocks.NewMockPeer("p2", peer2URL, 1001)
	peer3 = emocks.NewMockPeer("p3", peer3URL, 1002)
	peer4 = emocks.NewMockPeer("p4", peer4URL, 1003)
	peer5 = emocks.NewMockPeer("p5", peer5URL, 1004)
	peer6 = emocks.NewMockPeer("p6", peer6URL, 1005)

	allPeersWithState = []fab.Peer{peer1, peer2, peer3, peer4, peer5, peer6}

	// Peers with no block info
	peerA = mocks.NewMockPeer("pa", peer1URL)
	peerB = mocks.NewMockPeer("pb", peer2URL)

	allPeersWithoutState = []fab.Peer{peerA, peerB}
)

func TestBlockHeightPrioritySorter(t *testing.T) {
	threshold := 2
	sort := New(WithBlockHeightLagThreshold(threshold))

	for i := 0; i < 10; i++ {
		peers := sort(allPeersWithState)
		for i := 0; i < 3; i++ {
			p := peers[i]
			assert.Truef(t, p.URL() == peer6.URL() || p.URL() == peer5.URL() || p.URL() == peer4.URL(), "Unexpected peer [%s] in the top 3 with threshold %d", p.URL(), threshold)
		}
	}

	threshold = 3
	sort = New(WithBlockHeightLagThreshold(threshold))

	for i := 0; i < 10; i++ {
		peers := sort(allPeersWithState)
		for i := 0; i < 4; i++ {
			p := peers[i]
			assert.Truef(t, p.URL() == peer6.URL() || p.URL() == peer5.URL() || p.URL() == peer4.URL() || p.URL() == peer3.URL(), "Unexpected peer [%s] in the top 4 with threshold %d", p.URL(), threshold)
		}
	}
}

func TestBlockHeightPrioritySorterAny(t *testing.T) {
	sort := New(WithBlockHeightLagThreshold(Disable))

	top := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		top[sort(allPeersWithState)[0].URL()] = true
	}
	assert.Equalf(t, len(allPeersWithState), len(top), "Expecting each peer to have been selected first at least once")
}

func TestBlockHeightPrioritySorterNoBlockState(t *testing.T) {
	sort := New(WithBlockHeightLagThreshold(3))

	top := make(map[string]bool)
	for i := 0; i < 20; i++ {
		top[sort(allPeersWithoutState)[0].URL()] = true
	}
	assert.Equalf(t, len(allPeersWithoutState), len(top), "Expecting each peer to have been selected first at least once")
}

type priority int

func TestBlockHeightPrioritySorterDistributionRoundRobin(t *testing.T) {
	t.Run("All peers", func(t *testing.T) {
		// Round-robin is the default balancer
		sort := New(WithBlockHeightLagThreshold(Disable))

		iterations := len(allPeersWithState) * 10
		// Each peer should be called exactly the same number of times
		expected := iterations / len(allPeersWithState)

		count := make(map[string]int)

		for i := 0; i < iterations; i++ {
			peers := sort(allPeersWithState)
			p := peers[0]
			count[p.URL()] = count[p.URL()] + 1
		}

		for url, c := range count {
			assert.Truef(t, c == expected, "Expecting peer [%s] to have been first %d times but got only %d", url, expected, c)
		}
	})

	t.Run("With threshold 3", func(t *testing.T) {
		// Round-robin is the default balancer
		sort := New(WithBlockHeightLagThreshold(3))

		iterations := len(allPeersWithState) * 10

		// The top 4 peers should be called called exactly the same number of times and the
		// bottom 2 should be prioritized according to block height
		expectedTop := iterations / 4
		expectedBottom := iterations

		counts := make(map[priority]map[string]int)

		for p := 0; p < len(allPeersWithState); p++ {
			counts[priority(p)] = make(map[string]int)
		}

		for i := 0; i < iterations; i++ {
			peers := sort(allPeersWithState)
			for p, peer := range peers {
				counts[priority(p)][peer.URL()] = counts[priority(p)][peer.URL()] + 1
			}
		}

		for p := 0; p <= 3; p++ {
			checkCount(t, counts, priority(p), peer6URL, expectedTop)
		}

		checkCount(t, counts, priority(4), peer2URL, expectedBottom)
		checkCount(t, counts, priority(5), peer1URL, expectedBottom)
	})

}

func checkCount(t *testing.T, counts map[priority]map[string]int, p priority, peerURL string, expected int) {
	count := counts[p]
	actual := count[peerURL]
	assert.Truef(t, actual == expected, "Expecting peer [%s] to have been called %d times at priority %d but got only %d", peerURL, expected, p, actual)
}

func TestBlockHeightPrioritySorterDistributionRandom(t *testing.T) {
	sort := New(
		WithBlockHeightLagThreshold(5),
		WithBalancer(balancer.Random()),
	)

	iterations := 1000
	expectedMin := iterations/len(allPeersWithState) - 50

	count := make(map[string]int)

	for i := 0; i < iterations; i++ {
		peers := sort(allPeersWithState)
		p := peers[0]
		count[p.URL()] = count[p.URL()] + 1
	}

	for url, c := range count {
		assert.Truef(t, c >= expectedMin, "Expecting peer [%s] to have been first at least %d times but got only %d", url, expectedMin, c)
	}
}

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())
	os.Exit(m.Run())
}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"
	"unsafe"

	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

const (
	normalTimeout = 5 * time.Second

	normalSweepTime = 5 * time.Second
	normalIdleTime  = 10 * time.Second
	shortSweepTime  = 100 * time.Millisecond
	shortIdleTime   = 200 * time.Millisecond
	shortSleepTime  = 1000
)

func TestConnectorHappyPath(t *testing.T) {
	connector := NewCachingConnector(normalSweepTime, normalIdleTime)
	defer connector.Close()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn1, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	assert.NotEqual(t, connectivity.Connecting, conn1.GetState(), "connection should not be connecting")
	assert.NotEqual(t, connectivity.Shutdown, conn1.GetState(), "connection should not be shutdown")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn2, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")
	assert.Equal(t, unsafe.Pointer(conn1), unsafe.Pointer(conn2), "connections should match")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn3, err := connector.DialContext(ctx, endorserAddr[1], grpc.WithInsecure())
	cancel()

	assert.NotEqual(t, connectivity.Connecting, conn3.GetState(), "connection should not be connecting")
	assert.NotEqual(t, connectivity.Shutdown, conn3.GetState(), "connection should not be shutdown")

	assert.Nil(t, err, "DialContext should have succeeded")
	assert.NotEqual(t, unsafe.Pointer(conn1), unsafe.Pointer(conn3), "connections should not match")
}

func TestConnectorDoubleClose(t *testing.T) {
	connector := NewCachingConnector(normalSweepTime, normalIdleTime)
	defer connector.Close()
	connector.Close()
}

func TestConnectorHappyFlushNumber1(t *testing.T) {
	connector := NewCachingConnector(normalSweepTime, normalIdleTime)
	defer connector.Close()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn1, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	connector.Close()
	assert.Equal(t, connectivity.Shutdown, conn1.GetState(), "connection should be shutdown")
}

func TestConnectorHappyFlushNumber2(t *testing.T) {
	connector := NewCachingConnector(normalSweepTime, normalIdleTime)
	defer connector.Close()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn1, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn2, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn3, err := connector.DialContext(ctx, endorserAddr[1], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	connector.Close()
	assert.Equal(t, connectivity.Shutdown, conn1.GetState(), "connection should be shutdown")
	assert.Equal(t, connectivity.Shutdown, conn2.GetState(), "connection should be shutdown")
	assert.Equal(t, connectivity.Shutdown, conn3.GetState(), "connection should be shutdown")
}

func TestConnectorShouldJanitorRestart(t *testing.T) {
	connector := NewCachingConnector(shortSweepTime, shortIdleTime)
	defer connector.Close()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn1, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	connector.ReleaseConn(conn1)
	time.Sleep(shortIdleTime * 3)

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn2, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	assert.NotEqual(t, unsafe.Pointer(conn1), unsafe.Pointer(conn2), "connections should be different due to disconnect")
}

func TestConnectorShouldSweep(t *testing.T) {
	connector := NewCachingConnector(shortSweepTime, shortIdleTime)
	defer connector.Close()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn1, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn3, err := connector.DialContext(ctx, endorserAddr[1], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	connector.ReleaseConn(conn1)
	time.Sleep(shortIdleTime * 3)
	assert.Equal(t, connectivity.Shutdown, conn1.GetState(), "connection should be shutdown")
	assert.NotEqual(t, connectivity.Shutdown, conn3.GetState(), "connection should not be shutdown")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn4, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	assert.NotEqual(t, unsafe.Pointer(conn1), unsafe.Pointer(conn4), "connections should be different due to disconnect")
}

func TestConnectorConcurrent(t *testing.T) {
	const goroutines = 50

	connector := NewCachingConnector(shortSweepTime, shortIdleTime)
	defer connector.Close()

	wg := sync.WaitGroup{}

	// Test immediate release
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go testDial(t, &wg, connector, endorserAddr[i%2], 0, 1)
	}
	wg.Wait()

	// Test long intervals for releasing connection
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go testDial(t, &wg, connector, endorserAddr[i%2], shortSleepTime*3, 1)
	}
	wg.Wait()

	// Test mixed intervals for releasing connection
	wg.Add(goroutines)
	for i := 0; i < goroutines/2; i++ {
		go testDial(t, &wg, connector, endorserAddr[0], 0, 1)
		go testDial(t, &wg, connector, endorserAddr[1], shortSleepTime*3, 1)
	}
	wg.Wait()

	// Test random intervals for releasing connection
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go testDial(t, &wg, connector, endorserAddr[i%2], 0, shortSleepTime*3)
	}
	wg.Wait()
}

func testDial(t *testing.T, wg *sync.WaitGroup, connector *CachingConnector, addr string, minSleepBeforeRelease int, maxSleepBeforeRelease int) {
	defer wg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn, err := connector.DialContext(ctx, addr, grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	endorserClient := pb.NewEndorserClient(conn)
	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	proposal := pb.SignedProposal{}
	resp, err := endorserClient.ProcessProposal(context.Background(), &proposal)
	cancel()

	assert.Nil(t, err, "peer process proposal should not have error")
	assert.Equal(t, int32(200), resp.GetResponse().Status)

	randomSleep := rand.Intn(maxSleepBeforeRelease)
	time.Sleep(time.Duration(minSleepBeforeRelease)*time.Millisecond + time.Duration(randomSleep)*time.Millisecond)
	connector.ReleaseConn(conn)
}

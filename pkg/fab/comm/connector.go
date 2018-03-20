/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

const (
	connShutdownTimeout = 50 * time.Millisecond
)

// CachingConnector provides the ability to cache GRPC connections.
// It provides a GRPC compatible Context Dialer interface via the "DialContext" method.
// Connections provided by this component are monitored for becoming idle or entering shutdown state.
// When connections has its usages closed for longer than "idleTime", the connection is closed and removed
// from the connection cache. Callers must release connections by calling the "ReleaseConn" method.
// The Close method will flush all remaining open connections. This component should be considered
// unusable after calling Close.
//
// This component has been designed to be safe for concurrency.
type CachingConnector struct {
	conns         sync.Map
	sweepTime     time.Duration
	idleTime      time.Duration
	index         map[*grpc.ClientConn]*cachedConn
	lock          sync.Mutex
	waitgroup     sync.WaitGroup
	janitorChan   chan *cachedConn
	janitorDone   chan bool
	janitorClosed chan bool
}

type cachedConn struct {
	target    string
	conn      *grpc.ClientConn
	open      int
	lastOpen  time.Time
	lastClose time.Time
}

// NewCachingConnector creates a GRPC connection cache. The cache is governed by
// sweepTime and idleTime.
func NewCachingConnector(sweepTime time.Duration, idleTime time.Duration) *CachingConnector {
	cc := CachingConnector{
		conns:         sync.Map{},
		index:         map[*grpc.ClientConn]*cachedConn{},
		janitorChan:   make(chan *cachedConn),
		janitorDone:   make(chan bool),
		janitorClosed: make(chan bool, 1),
		sweepTime:     sweepTime,
		idleTime:      idleTime,
	}

	// cc.janitorClosed determines if a goroutine needs to be spun up.
	// The janitor is able to shut itself down when it has no connection to monitor.
	// When it shuts itself down, it pushes a value onto janitorClosed. We initialize
	// the go chan with a bootstrap value so that cachingConnector spins up the
	// goroutine on first usage.
	cc.janitorClosed <- true
	return &cc
}

// Close cleans up cached connections.
func (cc *CachingConnector) Close() {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	// Safety check to see if the connector has been closed. This represents a
	// bug in the calling code, but it's not good to panic here.
	if cc.janitorDone == nil {
		logger.Warn("Trying to close connector after already closed")
		return
	}
	logger.Debug("closing caching GRPC connector")

	select {
	case <-cc.janitorClosed:
		logger.Debugf("janitor not running")
	default:
		logger.Debugf("janitor running")
		cc.janitorDone <- true
		cc.waitgroup.Wait()
	}

	close(cc.janitorChan)
	close(cc.janitorClosed)
	close(cc.janitorDone)
	cc.janitorDone = nil
}

// DialContext is a wrapper for grpc.DialContext where connections are cached.
func (cc *CachingConnector) DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	logger.Debugf("DialContext: %s", target)

	c, ok := cc.loadConn(target)
	if !ok {
		createdConn, err := cc.createConn(ctx, target, opts...)
		if err != nil {
			return nil, errors.WithMessage(err, "connection creation failed")
		}
		c = createdConn
	}

	if err := cc.openConn(ctx, c); err != nil {
		return nil, errors.Errorf("dialing connection timed out [%s]", target)
	}
	return c.conn, nil
}

// ReleaseConn notifies the cache that the connection is no longer in use.
func (cc *CachingConnector) ReleaseConn(conn *grpc.ClientConn) {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	// Safety check to see if the connector has been closed. This represents a
	// bug in the calling code, but it's not good to panic here.
	if cc.janitorDone == nil {
		logger.Warn("Trying to release connection after connector closed")

		if conn.GetState() != connectivity.Shutdown {
			logger.Warn("Connection is not shutdown, trying to close ...")
			conn.Close()
		}
		return
	}

	cconn, ok := cc.index[conn]
	if !ok {
		logger.Warnf("connection not found [%p]", conn)
		return
	}
	logger.Debugf("ReleaseConn [%s]", cconn.target)

	if cconn.open > 0 {
		cconn.lastClose = time.Now()
		cconn.open--
	}

	cc.updateJanitor(cconn)
}

func (cc *CachingConnector) loadConn(target string) (*cachedConn, bool) {
	connRaw, ok := cc.conns.Load(target)
	if ok {
		c, ok := connRaw.(*cachedConn)
		if ok {
			if c.conn.GetState() != connectivity.Shutdown {
				logger.Debugf("using cached connection [%s: %p]", target, c)
				return c, true
			}
			cc.shutdownConn(c)
		}
	}
	return nil, false
}

func (cc *CachingConnector) createConn(ctx context.Context, target string, opts ...grpc.DialOption) (*cachedConn, error) {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	cconn, ok := cc.loadConn(target)
	if ok {
		return cconn, nil
	}

	logger.Debugf("creating connection [%s]", target)
	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "dialing peer failed")
	}

	logger.Debugf("storing connection [%s]", target)
	cconn = &cachedConn{
		target: target,
		conn:   conn,
	}
	cc.conns.Store(target, cconn)
	cc.index[conn] = cconn

	return cconn, nil
}

func (cc *CachingConnector) openConn(ctx context.Context, c *cachedConn) error {

	err := waitConn(ctx, c.conn, connectivity.Ready)
	if err != nil {
		return err
	}

	cc.lock.Lock()
	defer cc.lock.Unlock()
	c.open++
	c.lastOpen = time.Now()
	cc.updateJanitor(c)

	logger.Debugf("connection was opened [%s]", c.target)
	return nil
}

func waitConn(ctx context.Context, conn *grpc.ClientConn, targetState connectivity.State) error {
	for {
		state := conn.GetState()
		if state == targetState {
			break
		}
		if !conn.WaitForStateChange(ctx, state) {
			return errors.Wrap(ctx.Err(), "waiting for connection failed")
		}
	}
	return nil
}

func (cc *CachingConnector) shutdownConn(cconn *cachedConn) {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	logger.Debugf("connection was shutdown [%s]", cconn.target)
	cc.conns.Delete(cconn.target)
	delete(cc.index, cconn.conn)

	cconn.open = 0
	cconn.lastClose = time.Time{}

	cc.updateJanitor(cconn)
}

func (cc *CachingConnector) removeConn(target string) {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	logger.Debugf("removing connection [%s]", target)
	connRaw, ok := cc.conns.Load(target)
	if ok {
		c, ok := connRaw.(*cachedConn)
		if ok {
			delete(cc.index, c.conn)
			cc.conns.Delete(target)
			if err := c.conn.Close(); err != nil {
				logger.Debugf("unable to close connection [%s]", err)
			}
		}
	}
}

func (cc *CachingConnector) updateJanitor(c *cachedConn) {
	select {
	case <-cc.janitorClosed:
		logger.Debugf("janitor not started")
		cc.waitgroup.Add(1)
		go janitor(cc.sweepTime, cc.idleTime, &cc.waitgroup, cc.janitorChan, cc.janitorClosed, cc.janitorDone, cc.removeConn)
	default:
		logger.Debugf("janitor already started")
	}
	cClone := *c

	cc.janitorChan <- &cClone
}

// The janitor monitors open connections for shutdown state or extended non-usage.
// This component operates by running a sweep with a period determined by "sweepTime".
// When a connection returned the GRPC status connectivity.Shutdown or when the connection
// has its usages closed for longer than "idleTime", the connection is closed and the
// "connRemove" notifier is called.
//
// The caching connector:
//    pushes connection information via the "conn" go channel.
//    notifies the janitor of close by closing the "done" go channel.
//
// The janitor:
//    calls "connRemove" callback when closing a connection.
//    decrements the "wg" waitgroup when exiting.
//    writes to the "done" go channel when closing due to becoming empty.

type connRemoveNotifier func(target string)

func janitor(sweepTime time.Duration, idleTime time.Duration, wg *sync.WaitGroup, conn chan *cachedConn, close chan bool, done chan bool, connRemove connRemoveNotifier) {
	logger.Debugf("starting connection janitor")
	defer wg.Done()

	conns := map[string]*cachedConn{}
	ticker := time.NewTicker(sweepTime)
	for {
		select {
		case <-done:
			if len(conns) > 0 {
				logger.Debugf("flushing connection janitor with open connections [%d]", len(conns))
			} else {
				logger.Debugf("flushing connection janitor")
			}
			flush(conns)
			return
		case c := <-conn:
			cache(conns, c)
		case <-ticker.C:
			rm := sweep(conns, idleTime)
			for _, target := range rm {
				connRemove(target)
				delete(conns, target)
			}

			if len(conns) == 0 {
				logger.Debugf("closing connection janitor")
				close <- true
				return
			}
		}
	}
}

func cache(conns map[string]*cachedConn, updateConn *cachedConn) {

	c, ok := conns[updateConn.target]
	if ok && updateConn.lastClose.IsZero() && updateConn.conn.GetState() == connectivity.Shutdown {
		logger.Debugf("connection shutdown detected in connection janitor")
		// We need to remove the connection from sweep consideration immediately
		// since the connector has already removed it. Otherwise we can have a race
		// between shutdown and creating a connection concurrently.
		delete(conns, updateConn.target)
		return
	}

	if !ok {
		logger.Debugf("new connection in connection janitor")
	} else if c.conn != updateConn.conn {
		logger.Debugf("connection change in connection janitor")

		if err := c.conn.Close(); err != nil {
			logger.Debugf("unable to close connection [%s]", err)
		}

	} else {
		logger.Debugf("updating existing connection in connection janitor")
	}

	conns[updateConn.target] = updateConn
}

func flush(conns map[string]*cachedConn) {
	for _, c := range conns {
		logger.Debugf("connection janitor closing connection [%s]", c.target)
		closeConn(c.conn)
	}
}

func sweep(conns map[string]*cachedConn, idleTime time.Duration) []string {
	rm := make([]string, 0, len(conns))
	now := time.Now()
	for _, c := range conns {
		if c.open == 0 && now.After(c.lastClose.Add(idleTime)) {
			logger.Debugf("connection janitor closing connection [%s]", c.target)
			rm = append(rm, c.target)
		} else if c.conn.GetState() == connectivity.Shutdown {
			logger.Debugf("connection already closed [%s]", c.target)
			rm = append(rm, c.target)
		}
	}
	return rm
}

func closeConn(conn *grpc.ClientConn) {
	if err := conn.Close(); err != nil {
		logger.Debugf("unable to close connection [%s]", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), connShutdownTimeout)
	if err := waitConn(ctx, conn, connectivity.Shutdown); err != nil {
		logger.Debugf("unable to wait for connection close [%s]", err)
	}
	cancel()
}

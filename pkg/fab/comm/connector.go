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
	conns     map[string]*cachedConn
	sweepTime time.Duration
	idleTime  time.Duration
	index     map[*grpc.ClientConn]*cachedConn
	// lock protects concurrent access to the connection cache
	// it is held during create, load, release, and sweep connection
	// operations. Note: it is released during openConn, which is
	// the blocking part of the connection process.
	lock          sync.RWMutex
	waitgroup     sync.WaitGroup
	janitorDone   chan bool
	janitorClosed chan bool
}

type cachedConn struct {
	target    string
	conn      *grpc.ClientConn
	open      int
	lastClose time.Time
}

// NewCachingConnector creates a GRPC connection cache. The cache is governed by
// sweepTime and idleTime.
func NewCachingConnector(sweepTime time.Duration, idleTime time.Duration) *CachingConnector {
	cc := CachingConnector{
		conns:         map[string]*cachedConn{},
		index:         map[*grpc.ClientConn]*cachedConn{},
		janitorDone:   make(chan bool, 1),
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
	cc.lock.RLock()
	// Safety check to see if the connector has been closed. This represents a
	// bug in the calling code, but it's not good to panic here.
	if cc.janitorDone == nil {
		cc.lock.RUnlock()
		logger.Warn("Trying to close connector after already closed")
		return
	}
	cc.lock.RUnlock()
	logger.Debug("closing caching GRPC connector")

	select {
	case <-cc.janitorClosed:
		logger.Debug("janitor not running")
	default:
		logger.Debug("janitor running")
		cc.janitorDone <- true
		cc.waitgroup.Wait()
	}

	cc.lock.Lock()
	defer cc.lock.Unlock()

	if len(cc.index) > 0 {
		logger.Debugf("flushing connection cache with open connections [%d]", len(cc.index))
	} else {
		logger.Debug("flushing connection cache")
	}

	cc.flush()
	close(cc.janitorClosed)
	close(cc.janitorDone)
	cc.janitorDone = nil
}

// DialContext is a wrapper for grpc.DialContext where connections are cached.
func (cc *CachingConnector) DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	logger.Debugf("DialContext: %s", target)

	cc.lock.Lock()

	createdConn, err := cc.createConn(ctx, target, opts...)
	if err != nil {
		cc.lock.Unlock()
		return nil, errors.WithMessage(err, "connection creation failed")
	}
	c := createdConn

	cc.lock.Unlock()

	if err := cc.openConn(ctx, c); err != nil {
		cc.lock.Lock()
		setClosed(c)
		cc.removeConn(c)
		cc.lock.Unlock()
		return nil, errors.WithMessagef(err, "dialing connection on target [%s]", target)
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
			if err := conn.Close(); err != nil {
				logger.Warnf("conn close failed err %s", err)
			}
		}
		return
	}

	cconn, ok := cc.index[conn]
	if !ok {
		logger.Warnf("connection not found [%p]", conn)
		return
	}
	logger.Debugf("ReleaseConn [%s]", cconn.target)

	setClosed(cconn)

	cc.ensureJanitorStarted()
}

func (cc *CachingConnector) loadConn(target string) (*cachedConn, bool) {
	c, ok := cc.conns[target]
	if ok {
		if c.conn.GetState() != connectivity.Shutdown {
			logger.Debugf("using cached connection [%s: %p]", target, c)
			// Set connection open as soon as it is loaded to prevent the janitor
			// from sweeping it
			c.open++
			return c, true
		}
		cc.shutdownConn(c)
	}
	return nil, false
}

func (cc *CachingConnector) createConn(ctx context.Context, target string, opts ...grpc.DialOption) (*cachedConn, error) {
	if cc.janitorDone == nil {
		return nil, errors.New("caching connector is closed")
	}

	cconn, ok := cc.loadConn(target)
	if ok {
		return cconn, nil
	}

	logger.Debugf("creating connection [%s]", target)
	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "dialing node failed")
	}

	logger.Debugf("storing connection [%s]", target)
	cconn = &cachedConn{
		target: target,
		conn:   conn,
		open:   1,
	}

	cc.conns[target] = cconn
	cc.index[conn] = cconn

	return cconn, nil
}

func (cc *CachingConnector) openConn(ctx context.Context, c *cachedConn) error {

	err := waitConn(ctx, c.conn, connectivity.Ready)
	if err != nil {
		return err
	}

	cc.ensureJanitorStarted()

	logger.Debugf("connection was opened [%s]", c.target)
	return nil
}

func waitConn(ctx context.Context, conn *grpc.ClientConn, targetState connectivity.State) error {
	for {
		state := conn.GetState()
		if state == targetState {
			break
		}
		if state == connectivity.TransientFailure {
			// The server was probably restarted. It's better for the client to retry with a new connection rather
			// than reusing a cached connection that's in TRANSIENT_FAILURE state since it takes much longer to
			// recover while waiting for the state to change to READY - even if the server is up.
			return errors.Errorf("connection is in %s", state)
		}
		if !conn.WaitForStateChange(ctx, state) {
			return errors.Wrap(ctx.Err(), "waiting for connection failed")
		}
	}
	return nil
}

func (cc *CachingConnector) shutdownConn(cconn *cachedConn) {
	if cc.janitorDone == nil {
		logger.Debug("Connector already closed")
		return
	}

	logger.Debugf("connection was shutdown [%s]", cconn.target)
	delete(cc.conns, cconn.target)
	delete(cc.index, cconn.conn)

	cc.ensureJanitorStarted()
}

func (cc *CachingConnector) sweepAndRemove() {
	now := time.Now()
	for conn, cachedConn := range cc.index {
		if cachedConn.open == 0 && now.After(cachedConn.lastClose.Add(cc.idleTime)) {
			logger.Debugf("connection janitor closing connection [%s]", cachedConn.target)
			cc.removeConn(cachedConn)
		} else if conn.GetState() == connectivity.Shutdown {
			logger.Debugf("connection already closed [%s]", cachedConn.target)
			cc.removeConn(cachedConn)
		}
	}
}

func (cc *CachingConnector) removeConn(c *cachedConn) {
	logger.Debugf("removing connection [%s]", c.target)
	delete(cc.index, c.conn)
	delete(cc.conns, c.target)
	if err := c.conn.Close(); err != nil {
		logger.Debugf("unable to close connection [%s]", err)
	}
}

func (cc *CachingConnector) ensureJanitorStarted() {
	select {
	case <-cc.janitorClosed:
		logger.Debug("janitor not started")
		cc.waitgroup.Add(1)
		go cc.janitor()
	default:
	}
}

// janitor monitors open connections for shutdown state or extended non-usage.
// This component operates by running a sweep with a period determined by "sweepTime".
// When a connection returned the GRPC status connectivity.Shutdown or when the connection
// has its usages closed for longer than "idleTime", the connection is closed and the
// "connRemove" notifier is called.
//
// The caching connector:
//    notifies the janitor of close by closing the "done" go channel.
//
// The janitor:
//    calls "connRemove" callback when closing a connection.
//    decrements the "wg" waitgroup when exiting.
//    writes to the "done" go channel when closing due to becoming empty.
func (cc *CachingConnector) janitor() {
	logger.Debug("starting connection janitor")
	defer cc.waitgroup.Done()

	ticker := time.NewTicker(cc.sweepTime)
	defer ticker.Stop()
	for {
		select {
		case <-cc.janitorDone:
			return
		case <-ticker.C:
			cc.lock.Lock()
			cc.sweepAndRemove()
			numConn := len(cc.index)
			cc.lock.Unlock()
			if numConn == 0 {
				logger.Debug("closing connection janitor")
				cc.janitorClosed <- true
				return
			}
		}
	}
}

func (cc *CachingConnector) flush() {
	for _, c := range cc.index {
		logger.Debugf("flushing connection [%s]", c.target)
		closeConn(c.conn)
	}
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

func setClosed(cconn *cachedConn) {
	if cconn.open > 0 {
		cconn.lastClose = time.Now()
		cconn.open--
	}
}

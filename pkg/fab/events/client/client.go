/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	eventservice "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

// ConnectionState is the state of the client connection
type ConnectionState int32

const (
	// Disconnected indicates that the client is disconnected from the server
	Disconnected ConnectionState = iota
	// Connecting indicates that the client is in the process of establishing a connection
	Connecting
	// Connected indicates that the client is connected to the server
	Connected
)

// Client connects to an event server and receives events, such as block, filtered block,
// chaincode, and transaction status events. Client also monitors the connection to the
// event server and attempts to reconnect if the connection is closed.
type Client struct {
	*eventservice.Service
	params
	sync.RWMutex
	connEvent       chan *dispatcher.ConnectionEvent
	connectionState int32
	stopped         int32
	registerOnce    sync.Once
	afterConnect    handler
	beforeReconnect handler
}

type handler func() error

// New returns a new event client
func New(dispatcher eventservice.Dispatcher, opts ...options.Opt) *Client {
	params := defaultParams()
	options.Apply(params, opts)

	return &Client{
		Service:         eventservice.New(dispatcher, opts...),
		params:          *params,
		connectionState: int32(Disconnected),
	}
}

// SetAfterConnectHandler registers a handler that is called
// after the client connects to the event server. This allows for
// custom code to be executed for a particular
// event client implementation.
func (c *Client) SetAfterConnectHandler(h handler) {
	c.Lock()
	defer c.Unlock()
	c.afterConnect = h
}

func (c *Client) afterConnectHandler() handler {
	c.RLock()
	defer c.RUnlock()
	return c.afterConnect
}

// SetBeforeReconnectHandler registers a handler that will be called
// before retrying to reconnect to the event server. This allows for
// custom code to be executed for a particular event client implementation.
func (c *Client) SetBeforeReconnectHandler(h handler) {
	c.Lock()
	defer c.Unlock()
	c.beforeReconnect = h
}

func (c *Client) beforeReconnectHandler() handler {
	c.RLock()
	defer c.RUnlock()
	return c.beforeReconnect
}

// Connect connects to the peer and registers for events on a particular channel.
func (c *Client) Connect() error {
	if c.maxConnAttempts == 1 {
		return c.connect()
	}
	return c.connectWithRetry(c.maxConnAttempts, c.timeBetweenConnAttempts)
}

// CloseIfIdle closes the connection to the event server only if there are no outstanding
// registrations.
// Returns true if the client was closed. In this case the client may no longer be used.
// A return value of false indicates that the client could not be closed since
// there was at least one registration.
func (c *Client) CloseIfIdle() bool {
	logger.Debug("Attempting to close event client...")

	// Check if there are any outstanding registrations
	regInfoCh := make(chan *esdispatcher.RegistrationInfo)
	err := c.Submit(esdispatcher.NewRegistrationInfoEvent(regInfoCh))
	if err != nil {
		logger.Debugf("Submit failed %s", err)
		return false
	}
	regInfo := <-regInfoCh

	logger.Debugf("Outstanding registrations: %d", regInfo.TotalRegistrations)

	if regInfo.TotalRegistrations > 0 {
		logger.Debugf("Cannot stop client since there are %d outstanding registrations", regInfo.TotalRegistrations)
		return false
	}

	c.Close()

	return true
}

// Close closes the connection to the event server and releases all resources.
// Once this function is invoked the client may no longer be used.
func (c *Client) Close() {
	c.close(func() {
		c.Stop()
	})
}

// TransferRegistrations transfers all registrations into an EventSnapshot.
// The registrations are not closed and may susequently be transferred to a
// new event client.
// - close - if true then the client will also be closed
func (c *Client) TransferRegistrations(close bool) (fab.EventSnapshot, error) {
	if !close {
		return c.Transfer()
	}

	var snapshot fab.EventSnapshot
	var err error
	c.close(func() {
		logger.Debug("Stopping dispatcher and taking snapshot of all registrations...")
		snapshot, err = c.StopAndTransfer()
		if err != nil {
			logger.Errorf("An error occurred while stopping dispatcher and taking snapshot: %s", err)
		}
	})

	return snapshot, err
}

func (c *Client) close(stopHandler func()) {
	logger.Debug("Attempting to close event client...")

	if !c.setStoppped() {
		// Already stopped
		logger.Debug("Client already stopped")
		return
	}

	logger.Debug("Stopping client...")

	c.closeConnectEventChan()

	logger.Debug("Sending disconnect request...")

	errch := make(chan error)
	err1 := c.Submit(dispatcher.NewDisconnectEvent(errch))
	if err1 != nil {
		logger.Debugf("Submit failed %s", err1)
		return
	}
	err := <-errch

	if err != nil {
		logger.Warnf("Received error from disconnect request: %s", err)
	} else {
		logger.Debug("Received success from disconnect request")
	}

	logger.Debug("Stopping dispatcher...")

	stopHandler()

	c.mustSetConnectionState(Disconnected)

	logger.Debug("... event client is stopped")
}

func (c *Client) connect() error {
	if c.Stopped() {
		return errors.New("event client is closed")
	}

	if !c.setConnectionState(Disconnected, Connecting) {
		return errors.Errorf("unable to connect event client since client is [%s]. Expecting client to be in state [%s]", c.ConnectionState(), Disconnected)
	}

	logger.Debug("Submitting connection request...")

	errch := make(chan error, 1)
	err1 := c.Submit(dispatcher.NewConnectEvent(errch))
	if err1 != nil {
		return errors.Errorf("Submit failed %s", err1)
	}
	err := <-errch

	if err != nil {
		c.mustSetConnectionState(Disconnected)
		logger.Debugf("... got error in connection response: %s", err)
		return err
	}

	c.registerOnce.Do(func() {
		logger.Debug("Submitting connection event registration...")
		_, eventch, err1 := c.registerConnectionEvent()
		if err != nil {
			logger.Errorf("Error registering for connection events: %s", err1)
			c.Close()
		}
		c.connEvent = eventch
		go c.monitorConnection()
	})

	handlerImp := c.afterConnectHandler()
	if handlerImp != nil {
		err3 := c.t(handlerImp, errch)
		if err3 != nil {
			return err3
		}
	}

	c.setConnectionState(Connecting, Connected)

	logger.Debug("Submitting connected event")
	err2 := c.Submit(dispatcher.NewConnectedEvent())
	if err2 != nil {
		logger.Warnf("Submit failed %s", err2)
	}
	return err
}

func (c *Client) t(handlerImp handler, errch chan error) error {
	if err1 := handlerImp(); err1 != nil {
		logger.Warnf("Error invoking afterConnect handler: %s. Disconnecting...", err1)

		err2 := c.Submit(dispatcher.NewDisconnectEvent(errch))
		if err2 != nil {
			logger.Warnf("Submit failed %s", err2)
		}
		select {
		case disconnErr := <-errch:
			if disconnErr != nil {
				logger.Warnf("Received error from disconnect request: %s", disconnErr)
			} else {
				logger.Debug("Received success from disconnect request")
			}
		case <-time.After(c.respTimeout):
			logger.Warn("Timed out waiting for disconnect response")
		}

		c.setConnectionState(Connecting, Disconnected)

		return errors.WithMessage(err1, "error invoking afterConnect handler")
	}
	return nil
}

func (c *Client) connectWithRetry(maxAttempts uint, timeBetweenAttempts time.Duration) error {
	if c.Stopped() {
		return errors.New("event client is closed")
	}
	if timeBetweenAttempts < time.Second {
		timeBetweenAttempts = time.Second
	}

	var attempts uint
	for {
		if c.Stopped() {
			return errors.New("event client is closed")
		}

		attempts++
		logger.Debugf("Attempt #%d to connect...", attempts)
		if err := c.connect(); err != nil {
			logger.Warnf("... connection attempt failed: %s", err)
			if maxAttempts > 0 && attempts >= maxAttempts {
				logger.Warn("maximum connect attempts exceeded")
				return errors.New("maximum connect attempts exceeded")
			}
			time.Sleep(timeBetweenAttempts)
		} else {
			logger.Debug("... connect succeeded.")
			return nil
		}
	}
}

// RegisterBlockEvent registers for block events. If the client is not authorized to receive
// block events then an error is returned.
func (c *Client) RegisterBlockEvent(filter ...fab.BlockFilter) (fab.Registration, <-chan *fab.BlockEvent, error) {
	if !c.permitBlockEvents {
		return nil, nil, errors.New("block events are not permitted")
	}
	return c.Service.RegisterBlockEvent(filter...)
}

// registerConnectionEvent registers a connection event. The returned
// ConnectionEvent channel will be called whenever the client clients or disconnects
// from the event server
func (c *Client) registerConnectionEvent() (fab.Registration, chan *dispatcher.ConnectionEvent, error) {
	if c.Stopped() {
		return nil, nil, errors.New("event client is closed")
	}

	eventch := make(chan *dispatcher.ConnectionEvent, c.eventConsumerBufferSize)
	errch := make(chan error)
	regch := make(chan fab.Registration)
	err1 := c.Submit(dispatcher.NewRegisterConnectionEvent(eventch, regch, errch))
	if err1 != nil {
		return nil, nil, err1
	}
	select {
	case reg := <-regch:
		return reg, eventch, nil
	case err := <-errch:
		return nil, nil, err
	}
}

// Stopped returns true if the client has been stopped (disconnected)
// and is no longer usable.
func (c *Client) Stopped() bool {
	return atomic.LoadInt32(&c.stopped) == 1
}

func (c *Client) setStoppped() bool {
	return atomic.CompareAndSwapInt32(&c.stopped, 0, 1)
}

// ConnectionState returns the connection state
func (c *Client) ConnectionState() ConnectionState {
	return ConnectionState(atomic.LoadInt32(&c.connectionState))
}

// setConnectionState sets the connection state only if the given currentState
// matches the actual state. True is returned if the connection state was successfully set.
func (c *Client) setConnectionState(currentState, newState ConnectionState) bool {
	return atomic.CompareAndSwapInt32(&c.connectionState, int32(currentState), int32(newState))
}

func (c *Client) mustSetConnectionState(newState ConnectionState) {
	atomic.StoreInt32(&c.connectionState, int32(newState))
}

func (c *Client) monitorConnection() {
	logger.Debug("Monitoring connection")
	for event := range c.connEvent {
		if c.Stopped() {
			logger.Debugln("Event client has been stopped.")
			break
		}

		c.notifyConnectEventChan(event)

		if event.Connected {
			logger.Debug("Event client has connected")
		} else if c.reconn {
			logger.Warnf("Event client has disconnected. Details: %s", event.Err)
			if c.setConnectionState(Connected, Disconnected) {
				if event.Err.IsFatal() {
					logger.Warnf("Reconnect is not possible due to fatal error. Terminating: %s", event.Err)
					go c.Close()
					break
				}

				logger.Warn("Attempting to reconnect...")
				go c.reconnect()
			} else if c.setConnectionState(Connecting, Disconnected) {
				logger.Warn("Reconnect already in progress. Setting state to disconnected")
			}
		} else {
			logger.Debugf("Event client has disconnected. Terminating: %s", event.Err)
			go c.Close()
			break
		}
	}
	logger.Debug("Exiting connection monitor")
}

func (c *Client) reconnect() {
	logger.Debugf("Waiting %s before attempting to reconnect event client...", c.reconnInitialDelay)
	time.Sleep(c.reconnInitialDelay)

	logger.Debug("Attempting to reconnect event client...")

	handlerImp := c.beforeReconnectHandler()
	if handlerImp != nil {
		if err := handlerImp(); err != nil {
			logger.Errorf("Error invoking beforeReconnect handler: %s", err)
			return
		}
	}

	if err := c.connectWithRetry(c.maxReconnAttempts, c.timeBetweenConnAttempts); err != nil {
		logger.Warnf("Could not reconnect event client: %s", err)
		if !c.Stopped() {
			c.Close()
		}
	} else {
		logger.Infof("Event client has reconnected")
	}
}

func (c *Client) closeConnectEventChan() {
	c.Lock()
	defer c.Unlock()
	if c.connEventCh != nil {
		close(c.connEventCh)
	}
}

func (c *Client) notifyConnectEventChan(event *dispatcher.ConnectionEvent) {
	c.RLock()
	defer c.RUnlock()
	if c.connEventCh != nil {
		logger.Debugln("Sending connection event to subscriber.")
		c.connEventCh <- event
	}
}

func (s ConnectionState) String() string {
	switch s {
	case Disconnected:
		return "Disconnected"
	case Connected:
		return "Connected"
	case Connecting:
		return "Connecting"
	default:
		return "undefined"
	}
}

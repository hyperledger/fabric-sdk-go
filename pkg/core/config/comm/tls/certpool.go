/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package tls

import (
	"crypto/x509"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

var logger = logging.NewLogger("fabsdk/core")

// certPool is a thread safe wrapper around the x509 standard library
// cert pool implementation.
// It optionally allows loading the system trust store.
type certPool struct {
	useSystemCertPool bool
	certs             []*x509.Certificate
	certPool          *x509.CertPool
	certsByName       map[string][]int
	lock              sync.RWMutex
}

// NewCertPool new CertPool implementation
func NewCertPool(useSystemCertPool bool) fab.CertPool {
	return &certPool{
		useSystemCertPool: useSystemCertPool,
		certsByName:       make(map[string][]int),
		certPool:          x509.NewCertPool(),
	}
}

func (c *certPool) Get(certs ...*x509.Certificate) (*x509.CertPool, error) {
	c.lock.RLock()
	if len(certs) == 0 || c.containsCerts(certs...) {
		defer c.lock.RUnlock()
		return c.certPool, nil
	}
	c.lock.RUnlock()

	// We have a cert we have not encountered before, recreate the cert pool
	certPool, err := c.loadSystemCertPool()
	if err != nil {
		return nil, err
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	//add certs to SDK cert list
	for _, newCert := range certs {
		c.addCert(newCert)
	}
	//add all certs to cert pool
	for _, cert := range c.certs {
		certPool.AddCert(cert)
	}
	c.certPool = certPool

	return c.certPool, nil
}

func (c *certPool) addCert(newCert *x509.Certificate) {
	if newCert != nil && !c.containsCert(newCert) {
		n := len(c.certs)
		// Store cert
		c.certs = append(c.certs, newCert)
		// Store cert name index
		name := string(newCert.RawSubject)
		c.certsByName[name] = append(c.certsByName[name], n)
	}
}

func (c *certPool) containsCert(newCert *x509.Certificate) bool {
	possibilities := c.certsByName[string(newCert.RawSubject)]
	for _, p := range possibilities {
		if c.certs[p].Equal(newCert) {
			return true
		}
	}

	return false
}

func (c *certPool) containsCerts(certs ...*x509.Certificate) bool {
	for _, cert := range certs {
		if cert != nil && !c.containsCert(cert) {
			return false
		}
	}
	return true
}

func (c *certPool) loadSystemCertPool() (*x509.CertPool, error) {
	if !c.useSystemCertPool {
		return x509.NewCertPool(), nil
	}
	systemCertPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	logger.Debugf("Loaded system cert pool of size: %d", len(systemCertPool.Subjects()))

	return systemCertPool, nil
}

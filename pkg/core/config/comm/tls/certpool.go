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
	certsByName       map[string][]int
	lock              sync.RWMutex
}

// NewCertPool new CertPool implementation
func NewCertPool(useSystemCertPool bool) fab.CertPool {
	return &certPool{
		useSystemCertPool: useSystemCertPool,
		certsByName:       make(map[string][]int),
	}
}

func (c *certPool) Get(certs ...*x509.Certificate) (*x509.CertPool, error) {

	if len(certs) > 0 {
		c.lock.Lock()
		//add certs to SDK cert list
		for _, newCert := range certs {
			c.addCert(newCert)
		}
		c.lock.Unlock()
	}

	c.lock.RLock()
	defer c.lock.RUnlock()

	// create the cert pool
	certPool, err := c.loadSystemCertPool()
	if err != nil {
		return nil, err
	}

	//add all certs to cert pool
	for _, cert := range c.certs {
		certPool.AddCert(cert)
	}

	return certPool, nil
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

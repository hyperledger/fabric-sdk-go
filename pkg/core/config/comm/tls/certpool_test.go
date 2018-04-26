/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package tls

import (
	"crypto/x509"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var goodCert = &x509.Certificate{
	RawSubject: []byte("Good header"),
	Raw:        []byte("Good cert"),
}

func TestTLSCAConfig(t *testing.T) {
	tlsCertPool := NewCertPool(true).(*certPool)
	_, err := tlsCertPool.Get(goodCert)
	require.NoError(t, err)
	assert.Equal(t, true, tlsCertPool.useSystemCertPool)
	assert.NotNil(t, tlsCertPool.certPool)
	assert.NotNil(t, tlsCertPool.certsByName)

	originalLength := len(tlsCertPool.certs)
	//Try again with same cert
	_, err = tlsCertPool.Get(goodCert)
	assert.NoError(t, err, "TLS CA cert pool fetch failed")
	assert.False(t, len(tlsCertPool.certs) > originalLength, "number of certs in cert list shouldn't accept duplicates")

	// Test with system cert pool disabled
	tlsCertPool = NewCertPool(false).(*certPool)
	_, err = tlsCertPool.Get(goodCert)
	require.NoError(t, err)
	assert.Len(t, tlsCertPool.certs, 1)
	assert.Len(t, tlsCertPool.certPool.Subjects(), 1)
}

func TestTLSCAPoolManyCerts(t *testing.T) {
	size := 50

	tlsCertPool := NewCertPool(true).(*certPool)
	_, err := tlsCertPool.Get(goodCert)
	require.NoError(t, err)

	pool, err := tlsCertPool.Get()
	assert.NoError(t, err)
	originalLen := len(pool.Subjects())

	certs := createNCerts(size)
	pool, err = tlsCertPool.Get(certs[0])
	assert.NoError(t, err)
	assert.Len(t, pool.Subjects(), originalLen+1)

	pool, err = tlsCertPool.Get(certs...)
	assert.NoError(t, err)
	assert.Len(t, pool.Subjects(), originalLen+size)
}

func TestConcurrent(t *testing.T) {
	concurrency := 1000
	certs := createNCerts(concurrency)

	tlsCertPool := NewCertPool(false).(*certPool)

	writeDone := make(chan bool)
	readDone := make(chan bool)

	for i := 0; i < concurrency; i++ {
		go func(c *x509.Certificate) {
			_, err := tlsCertPool.Get(c)
			assert.NoError(t, err)
			writeDone <- true
		}(certs[i])
		go func() {
			_, err := tlsCertPool.Get()
			assert.NoError(t, err)
			readDone <- true
		}()
	}

	for i := 0; i < concurrency; i++ {
		select {
		case b := <-writeDone:
			assert.True(t, b)
		case <-time.After(time.Second * 10):
			t.Fatalf("Timed out waiting for write %d", i)
		}

		select {
		case b := <-readDone:
			assert.True(t, b)
		case <-time.After(time.Second * 10):
			t.Fatalf("Timed out waiting for read %d", i)
		}
	}

	assert.Len(t, tlsCertPool.certs, concurrency)
	assert.Len(t, tlsCertPool.certPool.Subjects(), concurrency)
}

func createNCerts(n int) []*x509.Certificate {
	var certs []*x509.Certificate
	for i := 0; i < n; i++ {
		cert := &x509.Certificate{
			RawSubject: []byte(strconv.Itoa(i)),
			Raw:        []byte(strconv.Itoa(i)),
		}
		certs = append(certs, cert)
	}
	return certs
}

func BenchmarkTLSCertPool(b *testing.B) {
	tlsCertPool := NewCertPool(true).(*certPool)

	for n := 0; n < b.N; n++ {
		tlsCertPool.Get()
	}
}

func BenchmarkTLSCertPoolSameCert(b *testing.B) {
	tlsCertPool := NewCertPool(true).(*certPool)

	for n := 0; n < b.N; n++ {
		tlsCertPool.Get(goodCert)
	}
}

func BenchmarkTLSCertPoolDifferentCert(b *testing.B) {
	tlsCertPool := NewCertPool(true).(*certPool)
	certs := createNCerts(b.N)

	for n := 0; n < b.N; n++ {
		tlsCertPool.Get(certs[n])
	}
}

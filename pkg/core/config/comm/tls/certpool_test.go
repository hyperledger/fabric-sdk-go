/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package tls

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
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

const (
	tlsCaOrg1 = `-----BEGIN CERTIFICATE-----
MIICSDCCAe+gAwIBAgIQVy95bDHyGiHPiW/hN7iCEzAKBggqhkjOPQQDAjB2MQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTEfMB0GA1UEAxMWdGxz
Y2Eub3JnMS5leGFtcGxlLmNvbTAeFw0xODA3MjUxNDQxMjJaFw0yODA3MjIxNDQx
MjJaMHYxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQH
Ew1TYW4gRnJhbmNpc2NvMRkwFwYDVQQKExBvcmcxLmV4YW1wbGUuY29tMR8wHQYD
VQQDExZ0bHNjYS5vcmcxLmV4YW1wbGUuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0D
AQcDQgAEMl8XK0Rpr514HXVut0MS/PX07l7gWeXGCQkl8T8LBuuSjGEkgSIuOwpf
VqQv4TwXH0A8zIBrtxY2/W3/ERhhC6NfMF0wDgYDVR0PAQH/BAQDAgGmMA8GA1Ud
JQQIMAYGBFUdJQAwDwYDVR0TAQH/BAUwAwEB/zApBgNVHQ4EIgQg+tqYPgAj39pQ
2EH0hxR4SbPOmDRCmwiDsaVIj7tXIFYwCgYIKoZIzj0EAwIDRwAwRAIgUJVxM/57
1WMfcy56D2zw6g9APP5Z3g+Qg/Y5cScstkgCIBj0JVuemNxiQWdXZ/Qhc6sh4m5d
ngzYatfQtNv3/+4V
-----END CERTIFICATE-----`

	tlsCaOrg2 = `-----BEGIN CERTIFICATE-----
MIICSDCCAe+gAwIBAgIQRAmchEVD9462610qy8BdfDAKBggqhkjOPQQDAjB2MQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzEZMBcGA1UEChMQb3JnMi5leGFtcGxlLmNvbTEfMB0GA1UEAxMWdGxz
Y2Eub3JnMi5leGFtcGxlLmNvbTAeFw0xODA3MjUxNDQxMjJaFw0yODA3MjIxNDQx
MjJaMHYxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQH
Ew1TYW4gRnJhbmNpc2NvMRkwFwYDVQQKExBvcmcyLmV4YW1wbGUuY29tMR8wHQYD
VQQDExZ0bHNjYS5vcmcyLmV4YW1wbGUuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0D
AQcDQgAErO2jmz8uCcrVAghsV0CqWCbp55apZ+4Sww01eYswbeNWsFkXLeUxoDYW
24ClPai2+hXe8djlV/0J9dQN9Lb05aNfMF0wDgYDVR0PAQH/BAQDAgGmMA8GA1Ud
JQQIMAYGBFUdJQAwDwYDVR0TAQH/BAUwAwEB/zApBgNVHQ4EIgQgV28IES/J0+fq
SZlYwCgztWVcEH4gwOvZw3g3y5J194wwCgYIKoZIzj0EAwIDRwAwRAIgYEcFrfpI
gd8ZaCY75B07c87C1FkMJqom3TrdyLbb39kCIHS9zZg6t/W/rmSG6rJlXxqS3RRh
10Y4jiCH6so41N9w
-----END CERTIFICATE-----`

	tlsOrdererCert = `-----BEGIN CERTIFICATE-----
MIICNjCCAdygAwIBAgIRAO47NS1d5RtzwWFIUwWpWT0wCgYIKoZIzj0EAwIwbDEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xFDASBgNVBAoTC2V4YW1wbGUuY29tMRowGAYDVQQDExF0bHNjYS5l
eGFtcGxlLmNvbTAeFw0xODA3MjUxNDQxMjJaFw0yODA3MjIxNDQxMjJaMGwxCzAJ
BgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1TYW4gRnJh
bmNpc2NvMRQwEgYDVQQKEwtleGFtcGxlLmNvbTEaMBgGA1UEAxMRdGxzY2EuZXhh
bXBsZS5jb20wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQEQRujXijEID810f9Y
09zBcqucKcz8G4nRgsbfuLZiLgy3Cq/TJsdgjAIvfqR56KtQaupS6tU8xPtLFIr5
Vr6oo18wXTAOBgNVHQ8BAf8EBAMCAaYwDwYDVR0lBAgwBgYEVR0lADAPBgNVHRMB
Af8EBTADAQH/MCkGA1UdDgQiBCCKGZZgFT855X2eDdYagwYvZB8rkWu7xMsRGvvm
bworhDAKBggqhkjOPQQDAgNIADBFAiEAnmp7VjUxVbfrKRXfpW3X2O27doYxb1Z9
xjm288m0ljoCID1CMTrMDZn8M/YYpPrw9WkS3n2clykUQeMxeMN8uUCj
-----END CERTIFICATE-----`
)

func TestTLSCAConfigWithMultipleCerts(t *testing.T) {

	//prepare 3 certs
	certOrg1, err := getCertFromPEMBytes([]byte(tlsCaOrg1))
	assert.Nil(t, err)
	assert.NotNil(t, certOrg1)

	certOrg2, err := getCertFromPEMBytes([]byte(tlsCaOrg2))
	assert.Nil(t, err)
	assert.NotNil(t, certOrg2)

	certOrderer, err := getCertFromPEMBytes([]byte(tlsOrdererCert))
	assert.Nil(t, err)
	assert.NotNil(t, certOrderer)

	//number of subjects in system cert pool
	c, err := loadSystemCertPool(true)
	assert.Nil(t, err)
	numberOfSubjects := len(c.Subjects())

	//create certpool instance
	tlsCertPool, err := NewCertPool(true)
	assert.Nil(t, err)

	//Empty cert pool with just system cert pool certs
	pool, err := tlsCertPool.Get()
	assert.Nil(t, err)
	verifyCertPoolInstance(t, pool, tlsCertPool, 0, 0, 0, numberOfSubjects, 0)

	//add 2 certs
	tlsCertPool.Add(certOrderer, certOrg1)
	verifyCertPoolInstance(t, pool, tlsCertPool, 0, 2, 2, numberOfSubjects, 1)
	pool, err = tlsCertPool.Get()
	assert.Nil(t, err)
	verifyCertPoolInstance(t, pool, tlsCertPool, 2, 2, 2, numberOfSubjects, 0)

	//add 1 existing cert, queue should be unchanged and dirty flag should be off
	tlsCertPool.Add(certOrg1)
	verifyCertPoolInstance(t, pool, tlsCertPool, 2, 2, 2, numberOfSubjects, 0)
	pool, err = tlsCertPool.Get()
	assert.Nil(t, err)
	verifyCertPoolInstance(t, pool, tlsCertPool, 2, 2, 2, numberOfSubjects, 0)

	//try again, add 1 existing cert, queue should be unchanged and dirty flag should be off
	tlsCertPool.Add(certOrg1)
	verifyCertPoolInstance(t, pool, tlsCertPool, 2, 2, 2, numberOfSubjects, 0)
	pool, err = tlsCertPool.Get()
	assert.Nil(t, err)
	verifyCertPoolInstance(t, pool, tlsCertPool, 2, 2, 2, numberOfSubjects, 0)

	//add 2 existing certs, queue should be unchanged and dirty flag should be off
	tlsCertPool.Add(certOrderer, certOrg1)
	verifyCertPoolInstance(t, pool, tlsCertPool, 2, 2, 2, numberOfSubjects, 0)
	pool, err = tlsCertPool.Get()
	assert.Nil(t, err)
	verifyCertPoolInstance(t, pool, tlsCertPool, 2, 2, 2, numberOfSubjects, 0)

	//add 3 certs, (2 existing + 1 new), queue should have one extra cert and dirty flag should be on
	tlsCertPool.Add(certOrderer, certOrg1, certOrg2)
	verifyCertPoolInstance(t, pool, tlsCertPool, 2, 3, 3, numberOfSubjects, 1)
	pool, err = tlsCertPool.Get()
	assert.Nil(t, err)
	verifyCertPoolInstance(t, pool, tlsCertPool, 3, 3, 3, numberOfSubjects, 0)

	//add all 3 existing certs, queue should be unchanged and dirty flag should be off
	tlsCertPool.Add(certOrderer, certOrg1, certOrg2)
	verifyCertPoolInstance(t, pool, tlsCertPool, 3, 3, 3, numberOfSubjects, 0)
	pool, err = tlsCertPool.Get()
	assert.Nil(t, err)
	verifyCertPoolInstance(t, pool, tlsCertPool, 3, 3, 3, numberOfSubjects, 0)

}

func verifyCertPoolInstance(t *testing.T, pool *x509.CertPool, fabPool CertPool, numberOfCertsInPool, numberOfCerts, numberOfCertsByName, numberOfSubjects int, dirty int32) {
	assert.NotNil(t, fabPool)
	tlsCertPool := fabPool.(*certPool)
	assert.Equal(t, dirty, tlsCertPool.dirty)
	assert.Equal(t, numberOfCerts, len(tlsCertPool.certs))
	assert.Equal(t, numberOfCertsByName, len(tlsCertPool.certsByName))
	assert.Equal(t, numberOfSubjects+numberOfCertsInPool, len(pool.Subjects()))
}

func TestAddingDuplicateCertsToPool(t *testing.T) {
	//prepare 3 certs
	certOrg1, err := getCertFromPEMBytes([]byte(tlsCaOrg1))
	assert.Nil(t, err)
	assert.NotNil(t, certOrg1)

	certOrg2, err := getCertFromPEMBytes([]byte(tlsCaOrg2))
	assert.Nil(t, err)
	assert.NotNil(t, certOrg2)

	certOrderer, err := getCertFromPEMBytes([]byte(tlsOrdererCert))
	assert.Nil(t, err)
	assert.NotNil(t, certOrderer)

	//number of subjects in system cert pool
	c, err := loadSystemCertPool(true)
	assert.Nil(t, err)
	numberOfSubjects := len(c.Subjects())

	//create certpool instance
	tlsCertPool, err := NewCertPool(true)
	assert.Nil(t, err)

	//Empty cert pool with just system cert pool certs
	pool, err := tlsCertPool.Get()
	assert.Nil(t, err)
	verifyCertPoolInstance(t, pool, tlsCertPool, 0, 0, 0, numberOfSubjects, 0)

	//add multiple certs with duplicate
	tlsCertPool.Add(certOrderer, certOrg1, certOrderer, certOrg1, certOrg1, certOrg1, certOrderer, certOrderer)
	verifyCertPoolInstance(t, pool, tlsCertPool, 0, 2, 2, numberOfSubjects, 1)
	pool, err = tlsCertPool.Get()
	assert.Nil(t, err)
	verifyCertPoolInstance(t, pool, tlsCertPool, 2, 2, 2, numberOfSubjects, 0)

	//add multiple certs with duplicate
	tlsCertPool.Add(certOrderer, certOrg1, certOrderer, certOrg1, certOrg2, certOrg2, certOrg2, certOrderer, certOrderer)
	verifyCertPoolInstance(t, pool, tlsCertPool, 2, 3, 3, numberOfSubjects, 1)
	pool, err = tlsCertPool.Get()
	assert.Nil(t, err)
	verifyCertPoolInstance(t, pool, tlsCertPool, 3, 3, 3, numberOfSubjects, 0)
}

func TestRemoveDuplicatesCerts(t *testing.T) {

	//prepare 3 certs
	certOrg1, err := getCertFromPEMBytes([]byte(tlsCaOrg1))
	assert.Nil(t, err)
	assert.NotNil(t, certOrg1)

	certOrg2, err := getCertFromPEMBytes([]byte(tlsCaOrg2))
	assert.Nil(t, err)
	assert.NotNil(t, certOrg2)

	certOrderer, err := getCertFromPEMBytes([]byte(tlsOrdererCert))
	assert.Nil(t, err)
	assert.NotNil(t, certOrderer)

	certs := removeDuplicates(certOrg1, certOrg2, certOrg1, certOrg1, certOrg1, certOrg1, certOrderer)
	assert.Equal(t, 3, len(certs))
	var hasCertOrg1, hasCertOrg2, hasOrdCert bool
	for _, c := range certs {
		if c.Subject.CommonName == "tlsca.org1.example.com" {
			hasCertOrg1 = true
		} else if c.Subject.CommonName == "tlsca.org2.example.com" {
			hasCertOrg2 = true
		} else if c.Subject.CommonName == "tlsca.example.com" {
			hasOrdCert = true
		}
	}
	assert.True(t, hasCertOrg1)
	assert.True(t, hasCertOrg2)
	assert.True(t, hasOrdCert)
}

func TestTLSCAConfig(t *testing.T) {
	fabCertPool, err := NewCertPool(true)
	require.NoError(t, err)

	tlsCertPool := fabCertPool.(*certPool)

	tlsCertPool.Add(goodCert)
	_, err = tlsCertPool.Get()
	require.NoError(t, err)
	assert.NotNil(t, tlsCertPool.certsByName)

	originalLength := len(tlsCertPool.certs)
	//Try again with same cert
	tlsCertPool.Add(goodCert)
	_, err = tlsCertPool.Get()
	assert.NoError(t, err, "TLS CA cert pool fetch failed")
	assert.False(t, len(tlsCertPool.certs) > originalLength, "number of certs in cert list shouldn't accept duplicates")

	// Test with system cert pool disabled
	fabCertPool, err = NewCertPool(false)
	require.NoError(t, err)
	tlsCertPool = fabCertPool.(*certPool)

	tlsCertPool.Add(goodCert)
	cPool, err := tlsCertPool.Get()
	require.NoError(t, err)
	assert.Len(t, tlsCertPool.certs, 1)
	assert.Len(t, cPool.Subjects(), 1)
}

func TestTLSCAPoolManyCerts(t *testing.T) {
	size := 50

	fabCertPool, err := NewCertPool(true)
	require.NoError(t, err)

	tlsCertPool := fabCertPool.(*certPool)
	tlsCertPool.Add(goodCert)
	_, err = tlsCertPool.Get()
	require.NoError(t, err)

	pool, err := tlsCertPool.Get()
	assert.NoError(t, err)
	originalLen := len(pool.Subjects())

	certs := createNCerts(size)
	tlsCertPool.Add(certs[0])
	pool, err = tlsCertPool.Get()
	assert.NoError(t, err)
	assert.Len(t, pool.Subjects(), originalLen+1)

	tlsCertPool.Add(certs...)
	pool, err = tlsCertPool.Get()
	assert.NoError(t, err)
	assert.Len(t, pool.Subjects(), originalLen+size)
}

func TestConcurrent(t *testing.T) {
	concurrency := 1000
	certs := createNCerts(concurrency)

	fabCertPool, err := NewCertPool(false)
	require.NoError(t, err)

	tlsCertPool := fabCertPool.(*certPool)

	systemCerts := len(tlsCertPool.certPool.Subjects())

	writeDone := make(chan bool)
	readDone := make(chan bool)

	for i := 0; i < concurrency; i++ {
		go func(c *x509.Certificate) {
			tlsCertPool.Add(c)
			_, err := tlsCertPool.Get()
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

	certPool, err := tlsCertPool.Get()
	assert.Len(t, tlsCertPool.certs, concurrency)
	require.NoError(t, err)
	assert.Len(t, certPool.Subjects(), concurrency+systemCerts)
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
	tlsCertPool, err := NewCertPool(true)
	require.NoError(b, err)

	for n := 0; n < b.N; n++ {
		tlsCertPool.Get()
	}
}

func BenchmarkTLSCertPoolSameCert(b *testing.B) {
	tlsCertPool, err := NewCertPool(true)
	require.NoError(b, err)

	for n := 0; n < b.N; n++ {
		tlsCertPool.Add(goodCert)
		tlsCertPool.Get()
	}
}

func BenchmarkTLSCertPoolDifferentCert(b *testing.B) {
	tlsCertPool, err := NewCertPool(true)
	require.NoError(b, err)

	certs := createNCerts(b.N)

	for n := 0; n < b.N; n++ {
		tlsCertPool.Add(certs[n])
		tlsCertPool.Get()
	}
}

func getCertFromPEMBytes(pemCerts []byte) (*x509.Certificate, error) {
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}

		return cert, nil
	}

	return nil, errors.New("empty cert bytes provided")
}

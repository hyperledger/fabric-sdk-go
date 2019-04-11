/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"bytes"
	"encoding/hex"
	"path/filepath"
	"strconv"
	"testing"

	"strings"

	"crypto/tls"

	"reflect"

	"crypto/x509"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/test/mockfab"
	"github.com/stretchr/testify/assert"
)

func TestTLSConfigErrorAddingCertificate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockfab.BadTLSClientMockConfig(mockCtrl)

	_, err := TLSConfig(mockfab.BadCert, "", config)
	if err == nil {
		t.Fatal("Expected failure adding invalid certificate")
	}

	if !strings.Contains(err.Error(), mockfab.ErrorMessage) {
		t.Fatalf("Expected error: %s", mockfab.ErrorMessage)
	}
}

func TestTLSConfigErrorFromClientCerts(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockfab.BadTLSClientMockConfig(mockCtrl)

	_, err := TLSConfig(mockfab.BadCert, "", config)

	if err == nil {
		t.Fatal("Expected failure from loading client certs")
	}

	if !strings.Contains(err.Error(), mockfab.ErrorMessage) {
		t.Fatalf("Expected error: %s", mockfab.ErrorMessage)
	}
}

func TestTLSConfigHappyPath(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	testCertPool := x509.NewCertPool()
	certs := createNCerts(1)
	testCertPool.AddCert(certs[0])

	config := mockfab.CustomMockConfig(mockCtrl, testCertPool)

	serverHostOverride := "servernamebeingoverriden"

	tlsConfig, err := TLSConfig(mockfab.GoodCert, serverHostOverride, config)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if tlsConfig.ServerName != serverHostOverride {
		t.Fatal("Incorrect server name!")
	}

	if tlsConfig.RootCAs != testCertPool {
		t.Fatal("Incorrect cert pool")
	}

	if len(tlsConfig.Certificates) != 1 {
		t.Fatal("Incorrect number of certs")
	}

	if !reflect.DeepEqual(tlsConfig.Certificates[0], mockfab.TLSCert) {
		t.Fatal("Certs do not match")
	}
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

func TestNoTlsCertHash(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mockfab.NewMockEndpointConfig(mockCtrl)

	config.EXPECT().TLSClientCerts().Return([]tls.Certificate{})

	tlsCertHash, err := TLSCertHash(config)
	assert.NotNil(t, tlsCertHash)
	assert.Nil(t, err)
}

func TestEmptyTlsCertHash(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mockfab.NewMockEndpointConfig(mockCtrl)

	emptyCert := tls.Certificate{}
	config.EXPECT().TLSClientCerts().Return([]tls.Certificate{emptyCert})

	tlsCertHash, err := TLSCertHash(config)
	assert.NotNil(t, tlsCertHash)
	assert.Nil(t, err)
}

func TestTlsCertHash(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mockfab.NewMockEndpointConfig(mockCtrl)

	cert, err := tls.LoadX509KeyPair(filepath.Join("testdata", "server.crt"), filepath.Join("testdata", "server.key"))
	if err != nil {
		t.Fatalf("Unexpected error loading cert %s", err)
	}

	config.EXPECT().TLSClientCerts().Return([]tls.Certificate{cert})
	tlsCertHash, err := TLSCertHash(config)
	assert.NotNil(t, tlsCertHash)
	assert.Nil(t, err)
	// openssl x509 -fingerprint -sha256 -in testdata/server.crt
	// SHA256 Fingerprint=0D:D5:90:B8:A5:0E:A6:04:3E:A8:75:16:BF:77:A8:FE:E7:C5:62:2D:4C:B3:CB:99:12:74:72:2A:D8:BA:B8:92
	expectedHash, err := hex.DecodeString("0DD590B8A50EA6043EA87516BF77A8FEE7C5622D4CB3CB991274722AD8BAB892")
	if err != nil {
		t.Fatalf("Unexpected error decoding cert fingerprint %s", err)
	}
	if !bytes.Equal(tlsCertHash, expectedHash) {
		t.Fatal("Cert hash calculated incorrectly")
	}
}

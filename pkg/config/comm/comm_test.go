/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"crypto/x509"
	"testing"

	"strings"

	"crypto/tls"

	"reflect"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
)

func TestTLSConfigEmptyCertPoolAndCertificate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	// nil cert pool
	config.EXPECT().TLSCACertPool("").Return(nil, nil)

	_, err := TLSConfig("", "", config)
	if err == nil {
		t.Fatal("Expected failure with nil cert pool")
	}

	// empty cert pool
	certPool := x509.NewCertPool()
	config.EXPECT().TLSCACertPool("").Return(certPool, nil)

	_, err = TLSConfig("", "", config)
	if err == nil {
		t.Fatal("Expected failure with empty cert pool")
	}
}

func TestTLSConfigErrorAddingCertificate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	// empty cert pool and invalid certificate
	certificate := "invalid certificate"
	errMsg := "Error adding certificate to cert pool"
	certPool := x509.NewCertPool()
	config.EXPECT().TLSCACertPool("").Return(certPool, nil)
	config.EXPECT().TLSCACertPool(certificate).Return(certPool, errors.Errorf(errMsg))

	_, err := TLSConfig(certificate, "", config)
	if err == nil {
		t.Fatal("Expected failure adding invalid certificate")
	}

	if !strings.Contains(err.Error(), errMsg) {
		t.Fatalf("Expected error: %s", errMsg)
	}
}

func TestTLSConfigErrorFromClientCerts(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	certificate := "testCertificate"
	errMsg := "Error loading client certs"
	certPool := x509.NewCertPool()
	config.EXPECT().TLSCACertPool("").Return(certPool, nil)
	config.EXPECT().TLSCACertPool(certificate).Return(certPool, nil)
	config.EXPECT().TLSClientCerts().Return(nil, errors.Errorf(errMsg))

	_, err := TLSConfig(certificate, "", config)
	if err == nil {
		t.Fatal("Expected failure from loading client certs")
	}

	if !strings.Contains(err.Error(), errMsg) {
		t.Fatalf("Expected error: %s", errMsg)
	}
}

func TestTLSConfigHappyPath(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	certificate := "testCertificate"
	emptyCert := tls.Certificate{}
	serverHostOverride := "servernamebeingoverriden"
	certPool := x509.NewCertPool()
	config.EXPECT().TLSCACertPool("").Return(certPool, nil)
	config.EXPECT().TLSCACertPool(certificate).Return(certPool, nil)
	config.EXPECT().TLSClientCerts().Return([]tls.Certificate{emptyCert}, nil)

	tlsConfig, err := TLSConfig(certificate, serverHostOverride, config)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if tlsConfig.ServerName != serverHostOverride {
		t.Fatal("Incorrect server name!")
	}

	if tlsConfig.RootCAs != certPool {
		t.Fatal("Incorrect cert pool")
	}

	if len(tlsConfig.Certificates) != 1 {
		t.Fatal("Incorrect number of certs")
	}

	if !reflect.DeepEqual(tlsConfig.Certificates[0], emptyCert) {
		t.Fatal("Certs do not match")
	}
}

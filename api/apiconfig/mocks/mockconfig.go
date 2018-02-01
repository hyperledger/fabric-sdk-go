/*
Copyright SecureKey Technologies Inc., Unchain B.V. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mock_apiconfig

import (
	tls "crypto/tls"
	x509 "crypto/x509"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/pkg/errors"
)

// GoodCert is a mock of a good certificate
var GoodCert = &x509.Certificate{Raw: []byte{0, 1, 2}}

// BadCert is a mock of a bad certificate
var BadCert = &x509.Certificate{Raw: []byte{1, 2}}

// TLSCert is a mock of a tls.Certificate{}
var TLSCert = tls.Certificate{Certificate: [][]byte{{3}, {4}}}

// CertPool is a mock of a *x509.CertPool
var CertPool = x509.NewCertPool()

// ErrorMessage is a mock error message
const ErrorMessage = "default error message"

// DefaultMockConfig returns a default mock config for testing
func DefaultMockConfig(mockCtrl *gomock.Controller) *MockConfig {
	config := NewMockConfig(mockCtrl)

	config.EXPECT().TLSCACertPool(GoodCert).Return(CertPool, nil).AnyTimes()
	config.EXPECT().TLSCACertPool(BadCert).Return(CertPool, errors.New(ErrorMessage)).AnyTimes()
	config.EXPECT().TLSCACertPool().Return(CertPool, nil).AnyTimes()
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5).AnyTimes()
	config.EXPECT().TLSClientCerts().Return([]tls.Certificate{TLSCert}, nil).AnyTimes()

	return config
}

// BadTLSClientMockConfig returns a mock config for testing with TLSClientCerts() that always returns an error
func BadTLSClientMockConfig(mockCtrl *gomock.Controller) *MockConfig {
	config := NewMockConfig(mockCtrl)

	config.EXPECT().TLSCACertPool(GoodCert).Return(CertPool, nil).AnyTimes()
	config.EXPECT().TLSCACertPool(BadCert).Return(CertPool, errors.New(ErrorMessage)).AnyTimes()
	config.EXPECT().TLSCACertPool().Return(CertPool, nil).AnyTimes()
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5).AnyTimes()
	config.EXPECT().TLSClientCerts().Return(nil, errors.Errorf(ErrorMessage)).AnyTimes()

	return config
}

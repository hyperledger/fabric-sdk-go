/*
Copyright SecureKey Technologies Inc., Unchain B.V. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mockfab

import (
	tls "crypto/tls"
	x509 "crypto/x509"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
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
func DefaultMockConfig(mockCtrl *gomock.Controller) *MockEndpointConfig {
	config := NewMockEndpointConfig(mockCtrl)

	config.EXPECT().TLSCACertPool().Return(&MockCertPool{CertPool: CertPool}).AnyTimes()

	config.EXPECT().Timeout(fab.PeerConnection).Return(time.Second * 5).AnyTimes()
	config.EXPECT().TLSClientCerts().Return([]tls.Certificate{TLSCert}).AnyTimes()

	return config
}

// CustomMockConfig returns a custom mock config with custom certpool for testing
func CustomMockConfig(mockCtrl *gomock.Controller, certPool *x509.CertPool) *MockEndpointConfig {
	config := NewMockEndpointConfig(mockCtrl)

	config.EXPECT().TLSCACertPool().Return(&MockCertPool{CertPool: certPool}).AnyTimes()

	config.EXPECT().Timeout(fab.PeerConnection).Return(time.Second * 5).AnyTimes()
	config.EXPECT().TLSClientCerts().Return([]tls.Certificate{TLSCert}).AnyTimes()

	return config
}

// BadTLSClientMockConfig returns a mock config for testing with TLSClientCerts() that always returns an error
func BadTLSClientMockConfig(mockCtrl *gomock.Controller) *MockEndpointConfig {
	config := NewMockEndpointConfig(mockCtrl)

	config.EXPECT().TLSCACertPool().Return(&MockCertPool{Err: errors.New(ErrorMessage)}).AnyTimes()
	config.EXPECT().Timeout(fab.PeerConnection).Return(time.Second * 5).AnyTimes()
	config.EXPECT().TLSClientCerts().Return(nil).AnyTimes()

	return config
}

//MockCertPool for unit tests to mock CertPool
type MockCertPool struct {
	CertPool *x509.CertPool
	Err      error
}

//Get mock implementation of fab CertPool.Get()
func (c *MockCertPool) Get() (*x509.CertPool, error) {
	return c.CertPool, c.Err
}

//Add mock impl of adding certs to cert pool queue
func (c *MockCertPool) Add(certs ...*x509.Certificate) {

}

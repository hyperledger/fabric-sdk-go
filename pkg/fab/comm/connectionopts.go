/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"crypto/x509"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"google.golang.org/grpc/keepalive"
)

type params struct {
	hostOverride    string
	certificate     *x509.Certificate
	keepAliveParams keepalive.ClientParameters
	failFast        bool
	connectTimeout  time.Duration
}

func defaultParams() *params {
	return &params{
		failFast:       true,
		connectTimeout: 3 * time.Second,
	}
}

// WithHostOverride sets the host name that will be used to resolve the TLS certificate
func WithHostOverride(value string) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(hostOverrideSetter); ok {
			setter.SetHostOverride(value)
		}
	}
}

// WithCertificate sets the X509 certificate used for the TLS connection
func WithCertificate(value *x509.Certificate) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(certificateSetter); ok {
			setter.SetCertificate(value)
		}
	}
}

// WithKeepAliveParams sets the GRPC keep-alive parameters
func WithKeepAliveParams(value keepalive.ClientParameters) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(keepAliveParamsSetter); ok {
			setter.SetKeepAliveParams(value)
		}
	}
}

// WithFailFast sets the GRPC fail-fast parameter
func WithFailFast(value bool) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(failFastSetter); ok {
			setter.SetFailFast(value)
		}
	}
}

// WithConnectTimeout sets the GRPC connection timeout
func WithConnectTimeout(value time.Duration) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(connectTimeoutSetter); ok {
			setter.SetConnectTimeout(value)
		}
	}
}

func (p *params) SetHostOverride(value string) {
	logger.Debugf("HostOverride: %s", value)
	p.hostOverride = value
}

func (p *params) SetCertificate(value *x509.Certificate) {
	logger.Debugf("Certificate: %s", value)
	p.certificate = value
}

func (p *params) SetKeepAliveParams(value keepalive.ClientParameters) {
	logger.Debugf("KeepAliveParams: %#v", value)
	p.keepAliveParams = value
}

func (p *params) SetFailFast(value bool) {
	logger.Debugf("FailFast: %t", value)
	p.failFast = value
}

func (p *params) SetConnectTimeout(value time.Duration) {
	logger.Debugf("ConnectTimeout: %s", value)
	p.connectTimeout = value
}

type hostOverrideSetter interface {
	SetHostOverride(value string)
}

type certificateSetter interface {
	SetCertificate(value *x509.Certificate)
}

type keepAliveParamsSetter interface {
	SetKeepAliveParams(value keepalive.ClientParameters)
}

type failFastSetter interface {
	SetFailFast(value bool)
}

type connectTimeoutSetter interface {
	SetConnectTimeout(value time.Duration)
}

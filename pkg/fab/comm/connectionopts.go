/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"context"
	"crypto/x509"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/spf13/cast"
	"google.golang.org/grpc/keepalive"
)

type params struct {
	hostOverride    string
	certificate     *x509.Certificate
	keepAliveParams keepalive.ClientParameters
	failFast        bool
	insecure        bool
	connectTimeout  time.Duration
	parentContext   context.Context
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

// WithParentContext sets the parent context
func WithParentContext(value context.Context) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(parentContextSetter); ok {
			setter.SetParentContext(value)
		}
	}
}

// WithInsecure indicates to fall back to an insecure connection if the
// connection URL does not specify a protocol
func WithInsecure() options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(insecureSetter); ok {
			setter.SetInsecure(true)
		}
	}
}

func (p *params) SetHostOverride(value string) {
	logger.Debugf("HostOverride: %s", value)
	p.hostOverride = value
}

func (p *params) SetCertificate(value *x509.Certificate) {
	if value != nil {
		logger.Debugf("setting certificate [subject: %s, serial: %s]", value.Subject, value.SerialNumber)
	} else {
		logger.Debug("setting nil certificate")
	}
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

func (p *params) SetInsecure(value bool) {
	logger.Debugf("Insecure: %t", value)
	p.insecure = value
}

func (p *params) SetParentContext(value context.Context) {
	logger.Debugf("Setting parent context")
	p.parentContext = value
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

type insecureSetter interface {
	SetInsecure(value bool)
}

type connectTimeoutSetter interface {
	SetConnectTimeout(value time.Duration)
}

type parentContextSetter interface {
	SetParentContext(value context.Context)
}

// OptsFromPeerConfig returns a set of connection options from the given peer config
func OptsFromPeerConfig(peerCfg *fab.PeerConfig) []options.Opt {

	opts := []options.Opt{
		WithHostOverride(getServerNameOverride(peerCfg)),
		WithFailFast(getFailFast(peerCfg)),
		WithKeepAliveParams(getKeepAliveOptions(peerCfg)),
		WithCertificate(peerCfg.TLSCACert),
	}
	if isInsecureAllowed(peerCfg) {
		opts = append(opts, WithInsecure())
	}

	return opts
}

func getServerNameOverride(peerCfg *fab.PeerConfig) string {
	if str, ok := peerCfg.GRPCOptions["ssl-target-name-override"].(string); ok {
		return str
	}
	return ""
}

func getFailFast(peerCfg *fab.PeerConfig) bool {
	if ff, ok := peerCfg.GRPCOptions["fail-fast"].(bool); ok {
		return cast.ToBool(ff)
	}
	return false
}

func getKeepAliveOptions(peerCfg *fab.PeerConfig) keepalive.ClientParameters {
	var kap keepalive.ClientParameters
	if kaTime, ok := peerCfg.GRPCOptions["keep-alive-time"]; ok {
		kap.Time = cast.ToDuration(kaTime)
	}
	if kaTimeout, ok := peerCfg.GRPCOptions["keep-alive-timeout"]; ok {
		kap.Timeout = cast.ToDuration(kaTimeout)
	}
	if kaPermit, ok := peerCfg.GRPCOptions["keep-alive-permit"]; ok {
		kap.PermitWithoutStream = cast.ToBool(kaPermit)
	}
	return kap
}

func isInsecureAllowed(peerCfg *fab.PeerConfig) bool {
	allowInsecure, ok := peerCfg.GRPCOptions["allow-insecure"].(bool)
	if ok {
		return allowInsecure
	}
	return false
}

/*
Copyright SecureKey Technologies Inc., Unchain B.V. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderer

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/urlutil"
	"github.com/spf13/cast"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// NewOrderer Returns a Orderer instance
// Deprecated: use orderer.New() instead
func NewOrderer(url string, certPath string, serverHostOverride string, config apiconfig.Config,
	kap keepalive.ClientParameters) (*Orderer, error) {
	var opts []grpc.DialOption

	timeout := config.TimeoutOrDefault(apiconfig.OrdererConnection)
	if kap.Time > 0 || kap.Timeout > 0 {
		opts = append(opts, grpc.WithKeepaliveParams(kap))
	}
	if urlutil.IsTLSEnabled(url) {
		certConfig := apiconfig.TLSConfig{Path: certPath}
		certificate, err := certConfig.TLSCert()

		if err != nil {
			return nil, err
		}

		tlsConfig, err := comm.TLSConfig(certificate, serverHostOverride, config)
		if err != nil {
			return nil, err
		}

		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	return &Orderer{url: urlutil.ToAddress(url), grpcDialOption: opts, dialTimeout: timeout}, nil
}

// NewOrdererFromConfig returns an Orderer instance constructed from orderer config
// Deprecated: use orderer.New() instead
func NewOrdererFromConfig(ordererCfg *apiconfig.OrdererConfig, config apiconfig.Config) (*Orderer, error) {

	serverHostOverride := ""
	if str, ok := ordererCfg.GRPCOptions["ssl-target-name-override"].(string); ok {
		serverHostOverride = str
	}

	var kap keepalive.ClientParameters
	if kaTime, ok := ordererCfg.GRPCOptions["keep-alive-time"]; ok {
		kap.Time = cast.ToDuration(kaTime)
	}
	if kaTimeout, ok := ordererCfg.GRPCOptions["keep-alive-timeout"]; ok {
		kap.Timeout = cast.ToDuration(kaTimeout)
	}
	if kaPermit, ok := ordererCfg.GRPCOptions["keep-alive-permit"]; ok {
		kap.PermitWithoutStream = cast.ToBool(kaPermit)
	}

	return NewOrderer(ordererCfg.URL, ordererCfg.TLSCACerts.Path, serverHostOverride, config, kap)
}

/*
Copyright SecureKey Technologies Inc., Unchain B.V. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderer

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/urlutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// NewOrderer Returns a Orderer instance
// Deprecated: use orderer.New() instead
func NewOrderer(url string, certPath string, serverHostOverride string, config apiconfig.Config) (*Orderer, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(config.TimeoutOrDefault(apiconfig.OrdererConnection)))
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
	return &Orderer{url: urlutil.ToAddress(url), grpcDialOption: opts}, nil
}

// NewOrdererFromConfig returns an Orderer instance constructed from orderer config
// Deprecated: use orderer.New() instead
func NewOrdererFromConfig(ordererCfg *apiconfig.OrdererConfig, config apiconfig.Config) (*Orderer, error) {

	serverHostOverride := ""
	if str, ok := ordererCfg.GRPCOptions["ssl-target-name-override"].(string); ok {
		serverHostOverride = str
	}

	return NewOrderer(ordererCfg.URL, ordererCfg.TLSCACerts.Path, serverHostOverride, config)
}

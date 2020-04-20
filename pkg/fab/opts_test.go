/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"crypto/tls"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	commtls "github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm/tls"
)

var (
	m0  = &EndpointConfig{}
	m1  = &mockTimeoutConfig{}
	m4  = &mockrderersConfig{}
	m5  = &mockOrdererConfig{}
	m6  = &mockPeersConfig{}
	m7  = &mockPeerConfig{}
	m8  = &mockNetworkConfig{}
	m9  = &mockNetworkPeers{}
	m10 = &mockChannelConfig{}
	m11 = &mockChannelPeers{}
	m12 = &mockChannelOrderers{}
	m13 = &mockTLSCACertPool{}
	m15 = &mockTLSClientCerts{}
	m16 = &mockCryptoConfigPath{}
)

func TestCreateCustomFullEndpointConfig(t *testing.T) {
	var opts []interface{}
	opts = append(opts, m0)
	// try to build with the overall interface (m0 is the overall interface implementation)
	endpointConfigOption, err := BuildConfigEndpointFromOptions(opts...)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}
	if endpointConfigOption == nil {
		t.Fatal("BuildConfigEndpointFromOptions call returned nil")
	}
}

func TestCreateCustomEndpointConfig(t *testing.T) {
	// try to build with partial interfaces
	endpointConfigOption, err := BuildConfigEndpointFromOptions(m1, m4, m5, m6, m7, m8, m9, m10)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}
	var eco *EndpointConfigOptions
	var ok bool
	if eco, ok = endpointConfigOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %T", endpointConfigOption)
	}
	if eco == nil {
		t.Fatal("build ConfigEndpointOption returned is nil")
	}
	tmout := eco.Timeout(fab.PeerConnection)
	if tmout < 0 {
		t.Fatalf("EndpointConfig was supposed to have Timeout function overridden from Options but was not %+v. Timeout: %s", eco, tmout)
	}

	// verify if an interface was not passed as an option but was not nil, it should be nil
	if eco.channelPeers != nil {
		t.Fatalf("channelPeers created with nil interface but got non nil one. %s", eco.channelPeers)
	}
}

func TestCreateCustomEndpointConfigRemainingFunctions(t *testing.T) {
	// test other sub interface functions
	endpointConfigOption, err := BuildConfigEndpointFromOptions(m11, m12, m13, m15, m16)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}
	var eco *EndpointConfigOptions
	var ok bool
	if eco, ok = endpointConfigOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %T", endpointConfigOption)
	}
	if eco == nil {
		t.Fatal("build ConfigEndpointOption returned is nil")
	}
	// verify that their functions are available
	p := eco.ChannelPeers("")
	if len(p) != 1 {
		t.Fatalf("ChannelPeers did not return expected interface value. Expected: 1 ChannelPeer, Received: %d", len(p))
	}

	c := eco.TLSClientCerts()
	if len(c) != 2 {
		t.Fatalf("TLSClientCerts did not return expected interface value. Expected: 2 Certificates, Received: %d", len(c))
	}

	// verify if an interface that was not passed as an option but was not nil, it should be nil
	if eco.timeout != nil {
		t.Fatalf("timeout created with nil timeout interface but got non nil one. %s", eco.timeout)
	}

	// now try with non related interface to test if an error returns
	var badType interface{}
	_, err = BuildConfigEndpointFromOptions(m12, m13, badType)
	if err == nil {
		t.Fatal("BuildConfigEndpointFromOptions did not return error with badType")
	}
}

func TestCreateCustomEndpointConfigWithSomeDefaultFunctions(t *testing.T) {
	// create a config with the first 7 interfaces to be overridden
	endpointConfigOption, err := BuildConfigEndpointFromOptions(m1, m4, m5, m6, m7)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}

	var eco *EndpointConfigOptions
	var ok bool
	if eco, ok = endpointConfigOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %T", endpointConfigOption)
	}
	if eco == nil {
		t.Fatal("build ConfigEndpointOption returned is nil")
	}

	// now inject default interfaces (using m0 as default interface for the sake of this test) for the ones that were not overridden by options above
	endpointConfigOptionWithSomeDefaults := UpdateMissingOptsWithDefaultConfig(eco, m0)

	// test if options updated interfaces with options are still working
	tmout := endpointConfigOptionWithSomeDefaults.Timeout(fab.PeerConnection)
	expectedTimeout := 10 * time.Second
	if tmout != expectedTimeout {
		t.Fatalf("EndpointConfig was supposed to have Timeout function overridden from Options but was not %+v. Timeout: [expected: %s, received: %s]", eco, expectedTimeout, tmout)
	}

	// now check if interfaces that are not updated are defaulted with m0
	if eco, ok = endpointConfigOptionWithSomeDefaults.(*EndpointConfigOptions); !ok {
		t.Fatalf("UpdateMissingOptsWithDefaultConfig did not return a Options instance %T", endpointConfigOptionWithSomeDefaults)
	}
	// cryptoConfigPath (m17) is among the interfaces that were not updated by options
	if eco.cryptoConfigPath == nil {
		t.Fatal("UpdateMissingOptsWithDefaultConfig did not set CryptoConfigPath() with default function implementation")
	}
	// tlsClientCerts (m16) is among the interfaces that were not updated by options
	if eco.tlsClientCerts == nil {
		t.Fatal("UpdateMissingOptsWithDefaultConfig did not set TLSClientCerts() with default function implementation")
	}
}

func TestIsEndpointConfigFullyOverridden(t *testing.T) {
	// test with the some interfaces
	endpointConfigOption, err := BuildConfigEndpointFromOptions(m1)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}

	var eco *EndpointConfigOptions
	var ok bool
	if eco, ok = endpointConfigOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %T", endpointConfigOption)
	}

	// test verify if some interfaces were not overridden according to BuildConfigEndpointFromOptions above,
	// only 3 interfaces were overridden, so expected value is false
	isFullyOverridden := IsEndpointConfigFullyOverridden(eco)
	if isFullyOverridden {
		t.Fatal("Expected not fully overridden EndpointConfig interface, but received fully overridden.")
	}

	// now try with no opts, expected value is also false
	endpointConfigOption, err = BuildConfigEndpointFromOptions()
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}
	if eco, ok = endpointConfigOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %T", endpointConfigOption)
	}

	isFullyOverridden = IsEndpointConfigFullyOverridden(eco)
	if isFullyOverridden {
		t.Fatal("Expected not fully overridden EndpointConfig interface, but received fully overridden.")
	}

	// now try with all opts, expected value is true this time
	endpointConfigOption, err = BuildConfigEndpointFromOptions(m1, m4, m5, m6, m7, m8, m9, m10, m11, m12, m13, m15, m16)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}
	if eco, ok = endpointConfigOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %T", endpointConfigOption)
	}

	isFullyOverridden = IsEndpointConfigFullyOverridden(eco)
	if !isFullyOverridden {
		t.Fatal("Expected fully overridden EndpointConfig interface, but received not fully overridden.")
	}
}

func TestCreateCustomEndpointConfigWithSomeDefaultFunctionsRemainingFunctions(t *testing.T) {
	// do the same test with the other interfaces in reverse
	endpointConfigOption, err := BuildConfigEndpointFromOptions(m8, m9, m10, m11, m12, m13, m15, m16)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}

	var eco *EndpointConfigOptions
	var ok bool
	if eco, ok = endpointConfigOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %T", endpointConfigOption)
	}
	if eco == nil {
		t.Fatal("build ConfigEndpointOption returned is nil")
	}

	// now inject default interfaces
	endpointConfigOptionWithSomeDefaults := UpdateMissingOptsWithDefaultConfig(eco, m0)

	//test that interfaces overridden by the options are still working
	m := endpointConfigOptionWithSomeDefaults.CryptoConfigPath()
	if m != "" {
		t.Fatalf("CryptoConfigPath did not return expected interface value. Expected: '%s', Received: %s", "", m)
	}
}

type mockTimeoutConfig struct{}

func (m *mockTimeoutConfig) Timeout(timeoutType fab.TimeoutType) time.Duration {
	return 10 * time.Second
}

type mockrderersConfig struct{}

func (m *mockrderersConfig) OrderersConfig() []fab.OrdererConfig {
	return []fab.OrdererConfig{{URL: "orderer1.com", GRPCOptions: nil, TLSCACert: nil}}
}

type mockOrdererConfig struct{}

func (m *mockOrdererConfig) OrdererConfig(name string) (*fab.OrdererConfig, bool, bool) {
	return &fab.OrdererConfig{URL: "o.com", GRPCOptions: nil, TLSCACert: nil}, true, false
}

type mockPeersConfig struct{}

func (m *mockPeersConfig) PeersConfig(org string) ([]fab.PeerConfig, bool) {
	return []fab.PeerConfig{{URL: "peer.com", GRPCOptions: nil, TLSCACert: nil}}, true
}

type mockPeerConfig struct{}

func (m *mockPeerConfig) PeerConfig(nameOrURL string) (*fab.PeerConfig, bool) {
	return &fab.PeerConfig{URL: "p.com", GRPCOptions: nil, TLSCACert: nil}, true
}

type mockNetworkConfig struct{}

func (m *mockNetworkConfig) NetworkConfig() *fab.NetworkConfig {
	return &fab.NetworkConfig{}
}

type mockNetworkPeers struct{}

func (m *mockNetworkPeers) NetworkPeers() []fab.NetworkPeer {
	return []fab.NetworkPeer{{PeerConfig: fab.PeerConfig{URL: "p.com", GRPCOptions: nil, TLSCACert: nil}, MSPID: ""}}
}

type mockChannelConfig struct{}

func (m *mockChannelConfig) ChannelConfig(name string) *fab.ChannelEndpointConfig {
	return &fab.ChannelEndpointConfig{}
}

type mockChannelPeers struct{}

func (m *mockChannelPeers) ChannelPeers(name string) []fab.ChannelPeer {
	return []fab.ChannelPeer{{}}
}

type mockChannelOrderers struct{}

func (m *mockChannelOrderers) ChannelOrderers(name string) []fab.OrdererConfig {
	return []fab.OrdererConfig{}
}

type mockTLSCACertPool struct{}

func (m *mockTLSCACertPool) TLSCACertPool() commtls.CertPool {
	return nil
}

type mockTLSClientCerts struct{}

func (m *mockTLSClientCerts) TLSClientCerts() []tls.Certificate {
	return []tls.Certificate{{}, {}}
}

type mockCryptoConfigPath struct{}

func (m *mockCryptoConfigPath) CryptoConfigPath() string {
	return ""
}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"crypto/tls"
	"crypto/x509"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
)

var (
	m0  = &EndpointConfig{}
	m1  = &mockTimeoutConfig{}
	m2  = &mockMspID{}
	m3  = &mockPeerMSPID{}
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
	m14 = &mockEventServiceType{}
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
		t.Fatalf("BuildConfigEndpointFromOptions call returned nil")
	}
}

func TestCreateCustomEndpointConfig(t *testing.T) {
	// try to build with partial interfaces
	endpointConfigOption, err := BuildConfigEndpointFromOptions(m1, m2, m3, m4, m5, m6, m7, m8, m9, m10)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}
	var eco *EndpointConfigOptions
	var ok bool
	if eco, ok = endpointConfigOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %T", endpointConfigOption)
	}
	if eco == nil {
		t.Fatalf("build ConfigEndpointOption returned is nil")
	}
	tmout := eco.Timeout(fab.EndorserConnection)
	if tmout < 0 {
		t.Fatalf("EndpointConfig was supposed to have Timeout function overridden from Options but was not %+v. Timeout: %s", eco, tmout)
	}
	m, err := eco.MSPID("")
	if err != nil {
		t.Fatalf("MSPID returned unexpected error %s", err)
	}
	if m != "testMSP" {
		t.Fatalf("MSPID did not return expected interface value. Expected: %s, Received: %s", "testMSP", m)
	}
	m, err = eco.PeerMSPID("")
	if err != nil {
		t.Fatalf("PeerMSPID returned unexpected error %s", err)
	}
	if m != "testPeerMSP" {
		t.Fatalf("MSPID did not return expected interface value. Expected: %s, Received: %s", "testPeerMSP", m)
	}

	// verify if an interface was not passed as an option but was not nil, it should be nil
	if eco.channelPeers != nil {
		t.Fatalf("channelPeers created with nil interface but got non nil one. %s", eco.channelPeers)
	}
}

func TestCreateCustomEndpointConfigRemainingFunctions(t *testing.T) {
	// test other sub interface functions
	endpointConfigOption, err := BuildConfigEndpointFromOptions(m11, m12, m13, m14, m15, m16)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}
	var eco *EndpointConfigOptions
	var ok bool
	if eco, ok = endpointConfigOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %T", endpointConfigOption)
	}
	if eco == nil {
		t.Fatalf("build ConfigEndpointOption returned is nil")
	}
	// verify that their functions are available
	p, err := eco.ChannelPeers("")
	if err != nil {
		t.Fatalf("ChannelPeers returned unexpected error %s", err)
	}
	if len(p) != 1 {
		t.Fatalf("ChannelPeers did not return expected interface value. Expected: 1 ChannelPeer, Received: %d", len(p))
	}

	c, err := eco.TLSClientCerts()
	if err != nil {
		t.Fatalf("TLSClientCerts returned unexpected error %s", err)
	}
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
		t.Fatalf("BuildConfigEndpointFromOptions did not return error with badType")
	}
}

func TestCreateCustomEndpointConfigWithSomeDefaultFunctions(t *testing.T) {
	// create a config with the first 7 interfaces to be overridden
	endpointConfigOption, err := BuildConfigEndpointFromOptions(m1, m2, m3, m4, m5, m6, m7)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}

	var eco *EndpointConfigOptions
	var ok bool
	if eco, ok = endpointConfigOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %T", endpointConfigOption)
	}
	if eco == nil {
		t.Fatalf("build ConfigEndpointOption returned is nil")
	}

	// now inject default interfaces (using m0 as default interface for the sake of this test) for the ones that were not overridden by options above
	endpointConfigOptionWithSomeDefaults := UpdateMissingOptsWithDefaultConfig(eco, m0)

	// test if options updated interfaces with options are still working
	tmout := endpointConfigOptionWithSomeDefaults.Timeout(fab.EndorserConnection)
	expectedTimeout := 10 * time.Second
	if tmout != expectedTimeout {
		t.Fatalf("EndpointConfig was supposed to have Timeout function overridden from Options but was not %+v. Timeout: [expected: %s, received: %s]", eco, expectedTimeout, tmout)
	}
	m, err := endpointConfigOptionWithSomeDefaults.MSPID("")
	if err != nil {
		t.Fatalf("MSPID returned unexpected error %s", err)
	}
	if m != "testMSP" {
		t.Fatalf("MSPID did not return expected interface value. Expected: %s, Received: %s", "testMSP", m)
	}

	// now check if interfaces that are not updated are defaulted with m0
	if eco, ok = endpointConfigOptionWithSomeDefaults.(*EndpointConfigOptions); !ok {
		t.Fatalf("UpdateMissingOptsWithDefaultConfig did not return a Options instance %T", endpointConfigOptionWithSomeDefaults)
	}
	// cryptoConfigPath (m17) is among the interfaces that were not updated by options
	if eco.cryptoConfigPath == nil {
		t.Fatalf("UpdateMissingOptsWithDefaultConfig did not set CryptoConfigPath() with default function implementation")
	}
	// tlsClientCerts (m16) is among the interfaces that were not updated by options
	if eco.tlsClientCerts == nil {
		t.Fatalf("UpdateMissingOptsWithDefaultConfig did not set TLSClientCerts() with default function implementation")
	}
}

func TestIsEndpointConfigFullyOverridden(t *testing.T) {
	// test with the some interfaces
	endpointConfigOption, err := BuildConfigEndpointFromOptions(m1, m2, m3)
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
		t.Fatalf("Expected not fully overridden EndpointConfig interface, but received fully overridden.")
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
		t.Fatalf("Expected not fully overridden EndpointConfig interface, but received fully overridden.")
	}

	// now try with all opts, expected value is true this time
	endpointConfigOption, err = BuildConfigEndpointFromOptions(m1, m2, m3, m4, m5, m6, m7, m8, m9, m10, m11, m12, m13, m14, m15, m16)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}
	if eco, ok = endpointConfigOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %T", endpointConfigOption)
	}

	isFullyOverridden = IsEndpointConfigFullyOverridden(eco)
	if !isFullyOverridden {
		t.Fatalf("Expected fully overridden EndpointConfig interface, but received not fully overridden.")
	}
}

func TestCreateCustomEndpointConfigWithSomeDefaultFunctionsRemainingFunctions(t *testing.T) {
	// do the same test with the other interfaces in reverse
	endpointConfigOption, err := BuildConfigEndpointFromOptions(m8, m9, m10, m11, m12, m13, m14, m15, m16)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}

	var eco *EndpointConfigOptions
	var ok bool
	if eco, ok = endpointConfigOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %T", endpointConfigOption)
	}
	if eco == nil {
		t.Fatalf("build ConfigEndpointOption returned is nil")
	}

	// now inject default interfaces
	endpointConfigOptionWithSomeDefaults := UpdateMissingOptsWithDefaultConfig(eco, m0)

	//test that interfaces overridden by the options are still working
	m := endpointConfigOptionWithSomeDefaults.CryptoConfigPath()
	if m != "" {
		t.Fatalf("CryptoConfigPath did not return expected interface value. Expected: '%s', Received: %s", "", m)
	}
	e := endpointConfigOptionWithSomeDefaults.EventServiceType()

	if e != fab.DeliverEventServiceType {
		t.Fatalf("MSPID did not return expected interface value. Expected: %d, Received: %d", fab.DeliverEventServiceType, e)

	}
}

type mockTimeoutConfig struct{}

func (m *mockTimeoutConfig) Timeout(timeoutType fab.TimeoutType) time.Duration {
	return 10 * time.Second
}

type mockMspID struct{}

func (m *mockMspID) MSPID(org string) (string, error) {
	return "testMSP", nil
}

type mockPeerMSPID struct{}

func (m *mockPeerMSPID) PeerMSPID(name string) (string, error) {
	return "testPeerMSP", nil
}

type mockrderersConfig struct{}

func (m *mockrderersConfig) OrderersConfig() ([]fab.OrdererConfig, error) {
	return []fab.OrdererConfig{{URL: "orderer1.com", GRPCOptions: nil, TLSCACerts: endpoint.TLSConfig{Path: "", Pem: ""}}}, nil
}

type mockOrdererConfig struct{}

func (m *mockOrdererConfig) OrdererConfig(name string) (*fab.OrdererConfig, error) {
	return &fab.OrdererConfig{URL: "o.com", GRPCOptions: nil, TLSCACerts: endpoint.TLSConfig{Path: "", Pem: ""}}, nil
}

type mockPeersConfig struct{}

func (m *mockPeersConfig) PeersConfig(org string) ([]fab.PeerConfig, error) {
	return []fab.PeerConfig{{URL: "peer.com", EventURL: "event.peer.com", GRPCOptions: nil, TLSCACerts: endpoint.TLSConfig{Path: "", Pem: ""}}}, nil
}

type mockPeerConfig struct{}

func (m *mockPeerConfig) PeerConfig(nameOrURL string) (*fab.PeerConfig, error) {
	return &fab.PeerConfig{URL: "p.com", EventURL: "event.p.com", GRPCOptions: nil, TLSCACerts: endpoint.TLSConfig{Path: "", Pem: ""}}, nil
}

type mockNetworkConfig struct{}

func (m *mockNetworkConfig) NetworkConfig() (*fab.NetworkConfig, error) {
	return &fab.NetworkConfig{}, nil
}

type mockNetworkPeers struct{}

func (m *mockNetworkPeers) NetworkPeers() ([]fab.NetworkPeer, error) {
	return []fab.NetworkPeer{{PeerConfig: fab.PeerConfig{URL: "p.com", EventURL: "event.p.com", GRPCOptions: nil, TLSCACerts: endpoint.TLSConfig{Path: "", Pem: ""}}, MSPID: ""}}, nil
}

type mockChannelConfig struct{}

func (m *mockChannelConfig) ChannelConfig(name string) (*fab.ChannelNetworkConfig, error) {
	return &fab.ChannelNetworkConfig{}, nil
}

type mockChannelPeers struct{}

func (m *mockChannelPeers) ChannelPeers(name string) ([]fab.ChannelPeer, error) {
	return []fab.ChannelPeer{{}}, nil
}

type mockChannelOrderers struct{}

func (m *mockChannelOrderers) ChannelOrderers(name string) ([]fab.OrdererConfig, error) {
	return []fab.OrdererConfig{}, nil
}

type mockTLSCACertPool struct{}

func (m *mockTLSCACertPool) TLSCACertPool(certConfig ...*x509.Certificate) (*x509.CertPool, error) {
	return nil, nil
}

type mockEventServiceType struct{}

func (m *mockEventServiceType) EventServiceType() fab.EventServiceType {
	return fab.DeliverEventServiceType
}

type mockTLSClientCerts struct{}

func (m *mockTLSClientCerts) TLSClientCerts() ([]tls.Certificate, error) {
	return []tls.Certificate{{}, {}}, nil
}

type mockCryptoConfigPath struct{}

func (m *mockCryptoConfigPath) CryptoConfigPath() string {
	return ""
}

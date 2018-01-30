/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"crypto/x509"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabca"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client"
	channelImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	identityImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/resource"
)

// FabricProvider represents the default implementation of Fabric objects.
type FabricProvider struct {
	providerContext apifabclient.ProviderContext
}

type fabContext struct {
	apifabclient.ProviderContext
	apifabclient.IdentityContext
}

// New creates a FabricProvider enabling access to core Fabric objects and functionality.
func New(ctx apifabclient.ProviderContext) *FabricProvider {
	f := FabricProvider{
		providerContext: ctx,
	}
	return &f
}

// NewResourceClient returns a new client initialized for the current instance of the SDK.
func (f *FabricProvider) NewResourceClient(ic apifabclient.IdentityContext) (apifabclient.Resource, error) {
	ctx := &fabContext{
		ProviderContext: f.providerContext,
		IdentityContext: ic,
	}
	client := clientImpl.New(ctx)

	return client, nil
}

// NewChannelClient returns a new client initialized for the current instance of the SDK.
//
// TODO - add argument with channel config interface (to enable channel configuration obtained from the network)
func (f *FabricProvider) NewChannelClient(ic apifabclient.IdentityContext, channelID string) (apifabclient.Channel, error) {
	ctx := &fabContext{
		ProviderContext: f.providerContext,
		IdentityContext: ic,
	}
	channel, err := channelImpl.New(ctx, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "NewChannel failed")
	}

	chOrderers, err := f.providerContext.Config().ChannelOrderers(channel.Name())
	if err != nil {
		return nil, errors.WithMessage(err, "reading channel orderers failed")
	}

	for _, ordererCfg := range chOrderers {

		orderer, err := orderer.New(f.providerContext.Config(), orderer.FromOrdererConfig(&ordererCfg))
		if err != nil {
			return nil, errors.WithMessage(err, "creating orderer failed")
		}
		err = channel.AddOrderer(orderer)
		if err != nil {
			return nil, errors.WithMessage(err, "adding orderer failed")
		}
	}

	return channel, nil
}

// NewEventHub initilizes the event hub.
func (f *FabricProvider) NewEventHub(ic apifabclient.IdentityContext, channelID string) (apifabclient.EventHub, error) {
	peerConfig, err := f.providerContext.Config().ChannelPeers(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "read configuration for channel peers failed")
	}

	var eventSource *apiconfig.ChannelPeer
	for _, p := range peerConfig {
		if p.EventSource && p.MspID == ic.MspID() {
			eventSource = &p
			break
		}
	}

	if eventSource == nil {
		return nil, errors.New("unable to find event source for channel")
	}

	// Event source found, create event hub
	eventCtx := events.Context{
		ProviderContext: f.providerContext,
		IdentityContext: ic,
	}
	return events.FromConfig(eventCtx, &eventSource.PeerConfig)
}

// NewChannelConfig initializes the channel config
func (f *FabricProvider) NewChannelConfig(ic apifabclient.IdentityContext, channelID string) (apifabclient.ChannelConfig, error) {

	ctx := chconfig.Context{
		ProviderContext: f.providerContext,
		IdentityContext: ic,
	}

	return chconfig.New(ctx, channelID)
}

// NewCAClient returns a new FabricCAClient initialized for the current instance of the SDK.
func (f *FabricProvider) NewCAClient(orgID string) (apifabca.FabricCAClient, error) {
	return fabricCAClient.NewFabricCAClient(orgID, f.providerContext.Config(), f.providerContext.CryptoSuite())
}

/////////////
// TODO - refactor the below (see if we really need to create these objects from the factory rather than directly)

// NewUser returns a new default implementation of a User.
func (f *FabricProvider) NewUser(name string, signingIdentity *apifabclient.SigningIdentity) (apifabclient.User, error) {

	user := identityImpl.NewUser(name, signingIdentity.MspID)

	user.SetPrivateKey(signingIdentity.PrivateKey)
	user.SetEnrollmentCertificate(signingIdentity.EnrollmentCert)

	return user, nil
}

// NewPeer returns a new default implementation of Peer
func (f *FabricProvider) NewPeer(url string, certificate *x509.Certificate, serverHostOverride string) (apifabclient.Peer, error) {
	return peerImpl.New(f.providerContext.Config(), peerImpl.WithURL(url), peerImpl.WithTLSCert(certificate), peerImpl.WithServerName(serverHostOverride))
}

// NewPeerFromConfig returns a new default implementation of Peer based configuration
func (f *FabricProvider) NewPeerFromConfig(peerCfg *apiconfig.NetworkPeer) (apifabclient.Peer, error) {
	return peerImpl.New(f.providerContext.Config(), peerImpl.FromPeerConfig(peerCfg))
}

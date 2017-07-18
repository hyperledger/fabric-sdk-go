/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabapi

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fabca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context"
)

// OrgContext currently represents the clients for an organization that the app is dealing with.
// TODO: better decription (e.g., possibility of holding discovery resources for the org & peers).
type OrgContext struct {
	mspClient fabca.FabricCAClient
}

// NewOrgContext creates a context based on the providers in the SDK
func NewOrgContext(factory context.OrgClientFactory, orgID string, config apiconfig.Config) (*OrgContext, error) {
	c := OrgContext{}

	// Initialize MSP client
	client, err := factory.NewMSPClient(orgID, config)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize MSP client [%s]", err)
	}
	c.mspClient = client

	return &c, nil
}

// MSPClient provides the MSP client of the context.
func (c *OrgContext) MSPClient() fabca.FabricCAClient {
	return c.mspClient
}

// Session represents an identity being used with clients.
// TODO: Better description.
type Session struct {
	factory context.SessionClientFactory
	user    fab.User
}

// NewSession creates a session from a context and a user (TODO)
func NewSession(user fab.User, factory context.SessionClientFactory) *Session {
	s := Session{
		factory: factory,
		user:    user,
	}

	return &s
}

// Identity returns the User in the session.
// TODO: reduce interface to idnetity
func (s *Session) Identity() fab.User {
	return s.user
}

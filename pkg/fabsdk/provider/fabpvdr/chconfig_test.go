// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/stretchr/testify/assert"

	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
)

func TestCreateChannelCfg(t *testing.T) {
	p := newInfraProvider(t)
	testChannelID := "test"
	p.chCfgCache = newMockChCfgCache(chconfig.NewChannelCfg(testChannelID))
	ctx := mocks.NewMockProviderContext()
	user := mspmocks.NewMockSigningIdentity("user", "user")
	clientCtx := &mockClientContext{
		Providers:       ctx,
		SigningIdentity: user,
	}

	m, err := p.CreateChannelCfg(clientCtx, "")
	assert.Nil(t, err)
	assert.NotNil(t, m)

	m, err = p.CreateChannelCfg(clientCtx, testChannelID)
	assert.Nil(t, err)
	assert.NotNil(t, m)

	p.Close()
}

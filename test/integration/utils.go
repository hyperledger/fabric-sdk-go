/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"

	api "github.com/hyperledger/fabric-sdk-go/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/util"
)

// GetOrdererAdmin ...
func GetOrdererAdmin(c api.FabricClient) (api.User, error) {
	keyDir := "ordererOrganizations/example.com/users/Admin@example.com/keystore"
	certDir := "ordererOrganizations/example.com/users/Admin@example.com/signcerts"
	return util.GetPreEnrolledUser(c, keyDir, certDir, "ordererAdmin")
}

// GetAdmin ...
func GetAdmin(c api.FabricClient, userOrg string) (api.User, error) {
	keyDir := fmt.Sprintf("peerOrganizations/%s.example.com/users/Admin@%s.example.com/keystore", userOrg, userOrg)
	certDir := fmt.Sprintf("peerOrganizations/%s.example.com/users/Admin@%s.example.com/signcerts", userOrg, userOrg)
	username := fmt.Sprintf("peer%sAdmin", userOrg)
	return util.GetPreEnrolledUser(c, keyDir, certDir, username)
}

// GetUser ...
func GetUser(c api.FabricClient, userOrg string) (api.User, error) {
	keyDir := fmt.Sprintf("peerOrganizations/%s.example.com/users/User1@%s.example.com/keystore", userOrg, userOrg)
	certDir := fmt.Sprintf("peerOrganizations/%s.example.com/users/User1@%s.example.com/signcerts", userOrg, userOrg)
	username := fmt.Sprintf("peer%sUser1", userOrg)
	return util.GetPreEnrolledUser(c, keyDir, certDir, username)
}

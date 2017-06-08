/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package integration

import (
	"fmt"

	fc "github.com/hyperledger/fabric-sdk-go/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/fabric-client/util"
)

// GetOrdererAdmin ...
func GetOrdererAdmin(c fc.Client) (fc.User, error) {
	keyDir := "ordererOrganizations/example.com/users/Admin@example.com/keystore"
	certDir := "ordererOrganizations/example.com/users/Admin@example.com/signcerts"
	return util.GetPreEnrolledUser(c, keyDir, certDir, "ordererAdmin")
}

// GetAdmin ...
func GetAdmin(c fc.Client, userOrg string) (fc.User, error) {
	keyDir := fmt.Sprintf("peerOrganizations/%s.example.com/users/Admin@%s.example.com/keystore", userOrg, userOrg)
	certDir := fmt.Sprintf("peerOrganizations/%s.example.com/users/Admin@%s.example.com/signcerts", userOrg, userOrg)
	username := fmt.Sprintf("peer%sAdmin", userOrg)
	return util.GetPreEnrolledUser(c, keyDir, certDir, username)
}

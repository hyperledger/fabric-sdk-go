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

package fabricclient

import (
	"github.com/hyperledger/fabric/bccsp"
)

// User ...
/**
 * The User struct represents users that have been enrolled and represented by
 * an enrollment certificate (ECert) and a signing key. The ECert must have
 * been signed by one of the CAs the blockchain network has been configured to trust.
 * An enrolled user (having a signing key and ECert) can conduct chaincode deployments,
 * transactions and queries with the Chain.
 *
 * User ECerts can be obtained from a CA beforehand as part of deploying the application,
 * or it can be obtained from the optional Fabric COP service via its enrollment process.
 *
 * Sometimes User identities are confused with Peer identities. User identities represent
 * signing capability because it has access to the private key, while Peer identities in
 * the context of the application/SDK only has the certificate for verifying signatures.
 * An application cannot use the Peer identity to sign things because the application doesn’t
 * have access to the Peer identity’s private key.
 *
 */
type User interface {
	GetName() string
	GetRoles() []string
	SetRoles([]string)
	GetEnrollmentCertificate() []byte
	SetEnrollmentCertificate(cert []byte)
	SetPrivateKey(privateKey bccsp.Key)
	GetPrivateKey() bccsp.Key
	GenerateTcerts(count int, attributes []string)
}

type user struct {
	name                  string
	roles                 []string
	PrivateKey            bccsp.Key // ****This key is temporary We use it to sign transaction until we have tcerts
	enrollmentCertificate []byte
}

// UserJSON ...
type UserJSON struct {
	PrivateKeySKI         []byte
	EnrollmentCertificate []byte
}

// NewUser ...
/**
 * Constructor for a user.
 *
 * @param {string} name - The user name
 */
func NewUser(name string) User {
	return &user{name: name}
}

// GetName ...
/**
 * Get the user name.
 * @returns {string} The user name.
 */
func (u *user) GetName() string {
	return u.name
}

// GetRoles ...
/**
 * Get the roles.
 * @returns {[]string} The roles.
 */
func (u *user) GetRoles() []string {
	return u.roles
}

// SetRoles ...
/**
 * Set the roles.
 * @param roles {[]string} The roles.
 */
func (u *user) SetRoles(roles []string) {
	u.roles = roles
}

// GetEnrollmentCertificate ...
/**
 * Returns the underlying ECert representing this user’s identity.
 */
func (u *user) GetEnrollmentCertificate() []byte {
	return u.enrollmentCertificate
}

// SetEnrollmentCertificate ...
/**
 * Set the user’s Enrollment Certificate.
 */
func (u *user) SetEnrollmentCertificate(cert []byte) {
	u.enrollmentCertificate = cert
}

// SetPrivateKey ...
/**
 * deprecated.
 */
func (u *user) SetPrivateKey(privateKey bccsp.Key) {
	u.PrivateKey = privateKey
}

// GetPrivateKey ...
/**
 * deprecated.
 */
func (u *user) GetPrivateKey() bccsp.Key {
	return u.PrivateKey
}

// GenerateTcerts ...
/**
 * Gets a batch of TCerts to use for transaction. there is a 1-to-1 relationship between
 * TCert and Transaction. The TCert can be generated locally by the SDK using the user’s crypto materials.
 * @param {int} count how many in the batch to obtain
 * @param {[]string} attributes  list of attributes to include in the TCert
 * @return {[]tcert} An array of TCerts
 */
func (u *user) GenerateTcerts(count int, attributes []string) {

}

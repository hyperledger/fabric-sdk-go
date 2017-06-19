/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

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

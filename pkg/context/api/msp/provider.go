/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

// Providers represents the MSP service providers context.
type Providers interface {
	IdentityManager(orgName string) (IdentityManager, bool)
}

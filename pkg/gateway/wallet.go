/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

// A Wallet stores identity information used to connect to a Hyperledger Fabric network.
// Instances are created using factory methods on the implementing objects.
type Wallet interface {
	Put(label string, id IdentityType) error
	Get(label string) (IdentityType, error)
	Remove(label string) error
	Exists(label string) bool
	List() []string
}

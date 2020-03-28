/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"strings"

	"github.com/pkg/errors"
)

const (
	// AccessDenied indicates that the user does not have permission to perform the operation
	AccessDenied = "access denied"
)

// DiscoveryError is an error originating at the Discovery service
type DiscoveryError struct {
	error
	target string
}

//Error returns string representation with target
func (e DiscoveryError) Error() string {
	return errors.Wrapf(e.error, "target [%s]", e.target).Error()
}

//Target returns url of the peer
func (e DiscoveryError) Target() string {
	return e.target
}

//IsAccessDenied checks if response contains access denied msg
func (e DiscoveryError) IsAccessDenied() bool {
	return strings.Contains(e.Error(), AccessDenied)
}

func newDiscoveryError(err error, target string) error {
	return DiscoveryError{target: target, error: err}
}

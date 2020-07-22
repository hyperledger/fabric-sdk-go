/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package persistence

import (
	"regexp"

	"github.com/pkg/errors"
)

// LabelRegexp is the regular expression controlling the allowed characters
// for the package label.
var LabelRegexp = regexp.MustCompile(`^[[:alnum:]][[:alnum:]_.+-]*$`)

// ValidateLabel return an error if the provided label contains any invalid
// characters, as determined by LabelRegexp.
func ValidateLabel(label string) error {
	if !LabelRegexp.MatchString(label) {
		return errors.Errorf("invalid label '%s'. Label must be non-empty, can only consist of alphanumerics, symbols from '.+-_', and can only begin with alphanumerics", label)
	}

	return nil
}

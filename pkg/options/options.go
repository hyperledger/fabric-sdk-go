/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package options

// Params represents a construct that holds
// a set of parameters
type Params interface{}

// Opt is an option that is applied to Params
type Opt func(opts Params)

// Apply applies the given options to the given Params
func Apply(params Params, opts []Opt) {
	for _, opt := range opts {
		opt(params)
	}
}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"reflect"
	"strings"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// GetProperties extracts the properties from the discovered peer.
func GetProperties(endpoint *discclient.Peer) fab.Properties {
	if endpoint.StateInfoMessage == nil {
		return nil
	}

	stateInfo := endpoint.StateInfoMessage.GetStateInfo()
	if stateInfo == nil || stateInfo.Properties == nil {
		return nil
	}

	properties := make(fab.Properties)

	val := reflect.ValueOf(stateInfo.Properties).Elem()

	for i := 0; i < val.NumField(); i++ {
		fType := val.Type().Field(i)

		// Exclude protobuf fields
		if !strings.HasPrefix(fType.Name, "XXX_") {
			properties[fType.Name] = val.Field(i).Interface()
		}
	}

	return properties
}

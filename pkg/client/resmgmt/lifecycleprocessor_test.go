/*
Copyright IntellectEU, NV. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	"testing"

	"github.com/golang/protobuf/proto"
	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalInOrderCCDefResults(t *testing.T) {
	t.Run("unmarshalls and sorts chaincodes by name", func(t *testing.T) {
		aCCDefRes := &lb.QueryChaincodeDefinitionsResult{
			ChaincodeDefinitions: []*lb.QueryChaincodeDefinitionsResult_ChaincodeDefinition{
				{Name: "a", Sequence: 1, Version: "v1"},
				{Name: "b", Sequence: 2, Version: "v5"},
				{Name: "c", Sequence: 2, Version: "v5"},
			},
		}

		bCCDefRes := &lb.QueryChaincodeDefinitionsResult{
			ChaincodeDefinitions: []*lb.QueryChaincodeDefinitionsResult_ChaincodeDefinition{},
		}

		//set cc definitions in reversed order
		for i := len(aCCDefRes.ChaincodeDefinitions) - 1; i >= 0; i-- {
			bCCDefRes.ChaincodeDefinitions = append(bCCDefRes.ChaincodeDefinitions, aCCDefRes.ChaincodeDefinitions[i])
		}

		aCCefResBytes, err := proto.Marshal(aCCDefRes)
		require.NoError(t, err)

		bCCDefResBytes, err := proto.Marshal(bCCDefRes)
		require.NoError(t, err)

		a, err := unmarshalInOrderCCDefResults("", aCCefResBytes)
		require.NoError(t, err)
		b, err := unmarshalInOrderCCDefResults("", bCCDefResBytes)
		require.NoError(t, err)

		assert.True(t, proto.Equal(a, b))
	})

	t.Run("fails to unmarshall a payload into QueryChaincodeDefinitionsResult", func(t *testing.T) {
		_, err := unmarshalInOrderCCDefResults("", []byte("some unknown payload"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal payload into QueryChaincodeDefinitionsResult")
	})

	t.Run("fails to unmarshall a payload into QueryChaincodeDefinitionResult", func(t *testing.T) {
		_, err := unmarshalInOrderCCDefResults("someName", []byte("some unknown payload"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal payload with name someName into QueryChaincodeDefinitionResult")
	})
}

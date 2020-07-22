/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	reqContext "context"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	"github.com/stretchr/testify/require"
)

func TestLifecycle_Install(t *testing.T) {
	lc := NewLifecycle()
	require.NotNil(t, lc)

	ctx := setupContext()

	reqCtx, cancel := contextImpl.NewRequest(ctx)
	defer cancel()

	t.Run("Success", func(t *testing.T) {
		resp, err := lc.Install(reqCtx, []byte("install package"), []fab.ProposalProcessor{&mocks.MockPeer{}})
		require.NoError(t, err)
		require.NotEmpty(t, resp)
	})

	t.Run("No package", func(t *testing.T) {
		resp, err := lc.Install(reqCtx, nil, []fab.ProposalProcessor{&mocks.MockPeer{}})
		require.EqualError(t, err, "chaincode package is required")
		require.Empty(t, resp)
	})

	t.Run("No targets", func(t *testing.T) {
		resp, err := lc.Install(reqCtx, []byte("install package"), nil)
		require.EqualError(t, err, "targets is required")
		require.Empty(t, resp)
	})

	t.Run("Marshal error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected marshal error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.protoMarshal = func(pb proto.Message) ([]byte, error) { return nil, errExpected }

		resp, err := lc.Install(reqCtx, []byte("install package"), []fab.ProposalProcessor{&mocks.MockPeer{}})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("Unmarshal error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected unmarshal error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.protoUnmarshal = func(buf []byte, pb proto.Message) error { return errExpected }

		resp, err := lc.Install(reqCtx, []byte("install package"), []fab.ProposalProcessor{&mocks.MockPeer{}})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("Context error", func(t *testing.T) {
		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.newContext = func(ctx reqContext.Context) (context.Client, bool) { return nil, false }

		resp, err := lc.Install(reqCtx, []byte("install package"), []fab.ProposalProcessor{&mocks.MockPeer{}})
		require.EqualError(t, err, "failed get client context from reqContext for txn header")
		require.Empty(t, resp)
	})

	t.Run("Txn Header error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected Txn Header error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.newTxnHeader = func(ctx context.Client, channelID string, opts ...fab.TxnHeaderOpt) (*txn.TransactionHeader, error) {
			return nil, errExpected
		}

		resp, err := lc.Install(reqCtx, []byte("install package"), []fab.ProposalProcessor{&mocks.MockPeer{}})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})
}

func TestLifecycle_GetInstalledPackage(t *testing.T) {
	lc := NewLifecycle()
	require.NotNil(t, lc)

	ctx := setupContext()

	reqCtx, cancel := contextImpl.NewRequest(ctx)
	defer cancel()

	t.Run("Success", func(t *testing.T) {
		resp, err := lc.GetInstalledPackage(reqCtx, "packageid", &mocks.MockPeer{})
		require.NoError(t, err)
		require.Empty(t, resp)
	})

	t.Run("Marshal error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected marshal error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.protoMarshal = func(pb proto.Message) ([]byte, error) { return nil, errExpected }

		resp, err := lc.GetInstalledPackage(reqCtx, "packageid", &mocks.MockPeer{})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("Unmarshal error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected unmarshal error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.protoUnmarshal = func(buf []byte, pb proto.Message) error { return errExpected }

		resp, err := lc.GetInstalledPackage(reqCtx, "packageid", &mocks.MockPeer{})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("Context error", func(t *testing.T) {
		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.newContext = func(ctx reqContext.Context) (context.Client, bool) { return nil, false }

		resp, err := lc.GetInstalledPackage(reqCtx, "packageid", &mocks.MockPeer{})
		require.EqualError(t, err, "failed get client context from reqContext for txn header")
		require.Empty(t, resp)
	})

	t.Run("Txn Header error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected Txn Header error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.newTxnHeader = func(ctx context.Client, channelID string, opts ...fab.TxnHeaderOpt) (*txn.TransactionHeader, error) {
			return nil, errExpected
		}

		resp, err := lc.GetInstalledPackage(reqCtx, "packageid", &mocks.MockPeer{})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})
}

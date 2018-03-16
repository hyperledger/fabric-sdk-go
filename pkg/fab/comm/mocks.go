/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"context"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"google.golang.org/grpc"
)

// MockCommManager is a non-caching comm manager used
// for unit testing
type MockCommManager struct {
}

// DialContext creates a connection
func (m *MockCommManager) DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, target, opts...)
}

// ReleaseConn closes the connection
func (m *MockCommManager) ReleaseConn(conn *grpc.ClientConn) {
	if err := conn.Close(); err != nil {
		logger.Warnf("Error closing connection: %s", err)
	}
}

// MockInfraProvider overrides the comm manager to return
// the MockCommManager
type MockInfraProvider struct {
	fabmocks.MockInfraProvider
}

// NewMockInfraProvider return a new MockInfraProvider
func NewMockInfraProvider() *MockInfraProvider {
	return &MockInfraProvider{}
}

// CommManager returns the MockCommManager
func (f *MockInfraProvider) CommManager() fab.CommManager {
	return &MockCommManager{}
}

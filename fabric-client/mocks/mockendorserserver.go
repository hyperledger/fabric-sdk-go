/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mocks

import (
	"golang.org/x/net/context"

	pb "github.com/hyperledger/fabric/protos/peer"
)

// MockEndorserServer mock endoreser server to process endorsement proposals
type MockEndorserServer struct {
	ProposalError error
}

// ProcessProposal mock implementation that returns success if error is not set
// error if it is
func (m *MockEndorserServer) ProcessProposal(context context.Context,
	proposal *pb.SignedProposal) (*pb.ProposalResponse, error) {
	if m.ProposalError == nil {
		return &pb.ProposalResponse{Response: &pb.Response{
			Status: 200,
		}}, nil
	}
	return &pb.ProposalResponse{Response: &pb.Response{
		Status:  500,
		Message: m.ProposalError.Error(),
	}}, m.ProposalError
}

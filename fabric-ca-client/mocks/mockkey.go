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

import "github.com/hyperledger/fabric/bccsp"

// MockKey mocks BCCSP key
type MockKey struct {
}

// Bytes ...
func (m *MockKey) Bytes() ([]byte, error) {
	return []byte("Not implemented"), nil
}

// SKI ...
func (m *MockKey) SKI() []byte {
	return []byte("Not implemented")
}

// Symmetric ...
func (m *MockKey) Symmetric() bool {
	return false
}

// Private ...
func (m *MockKey) Private() bool {
	return true
}

// PublicKey ...
func (m *MockKey) PublicKey() (bccsp.Key, error) {
	return m, nil
}

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

package keyvaluestore

import (
	"testing"
)

func TestFKVSMethods(t *testing.T) {
	stateStore, err := CreateNewFileKeyValueStore("/tmp/keyvaluestore")
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore return error[%s]", err)
	}
	stateStore.SetValue("testvalue", []byte("data"))
	value, err := stateStore.GetValue("testvalue")
	if err != nil {
		t.Fatalf("stateStore.GetValue return error[%s]", err)
	}
	if string(value) != "data" {
		t.Fatalf("stateStore.GetValue didn't return the right value")
	}

}

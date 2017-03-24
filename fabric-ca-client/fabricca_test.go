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

package fabricca

import (
	"testing"
)

func TestEnrollWithMissingParameters(t *testing.T) {
	fabricCAClient, err := NewFabricCAClient()
	if err != nil {
		t.Fatalf("NewFabricCAClient return error: %v", err)
	}
	_, _, err = fabricCAClient.Enroll("", "user1")
	if err == nil {
		t.Fatalf("Enroll didn't return error")
	}
	if err.Error() != "enrollmentID is empty" {
		t.Fatalf("Enroll didn't return right error")
	}
	_, _, err = fabricCAClient.Enroll("test", "")
	if err == nil {
		t.Fatalf("Enroll didn't return error")
	}
	if err.Error() != "enrollmentSecret is empty" {
		t.Fatalf("Enroll didn't return right error")
	}
}

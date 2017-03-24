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

// KeyValueStore ...
/**
 * Abstract class for a Key-Value store. The Chain class uses this store
 * to save sensitive information such as authenticated user's private keys,
 * certificates, etc.
 *
 */
type KeyValueStore interface {
	/**
	 * Get the value associated with name.
	 *
	 * @param {string} name of the key
	 * @returns {[]byte}
	 */
	GetValue(key string) ([]byte, error)

	/**
	 * Set the value associated with name.
	 * @param {string} name of the key to save
	 * @param {[]byte} value to save
	 */
	SetValue(key string, value []byte) error
}

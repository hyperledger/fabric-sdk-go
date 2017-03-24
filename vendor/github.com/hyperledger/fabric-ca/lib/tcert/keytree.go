/*
Copyright IBM Corp. 2016 All Rights Reserved.

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

package tcert

import (
	"fmt"
	"strings"

	"github.com/hyperledger/fabric/bccsp"
)

/*
 * A key tree is a hierarchy of derived keys with a single root key.
 * Each node in the tree has a key and a name, where the key is secret
 * and the name may be public.  If the secret associated with a node
 * is known, then the secret of each node in it's sub-tree can be derived
 * if the name of the nodes are known; however, it is not possible to
 * derive the keys associated with other nodes in the tree which are not
 * part of this sub-tree.
 *
 * This data structure is useful to support releasing a secret associated
 * with any node to an auditor without giving the auditor access to all
 * nodes in the tree.
 */

const (
	keyPathSep = "/"
)

// NewKeyTree is the constructor for a key tree
func NewKeyTree(bccspMgr bccsp.BCCSP, rootKey bccsp.Key) *KeyTree {
	tree := new(KeyTree)
	tree.bccspMgr = bccspMgr
	tree.rootKey = rootKey
	tree.keys = make(map[string]bccsp.Key)
	return tree
}

// KeyTree is a tree of derived keys
type KeyTree struct {
	bccspMgr bccsp.BCCSP
	rootKey  bccsp.Key
	keys     map[string]bccsp.Key
}

// GetKey returns a key associated with a specific path in the tree.
func (m *KeyTree) GetKey(path []string) (bccsp.Key, error) {
	if path == nil || len(path) == 0 {
		return m.rootKey, nil
	}
	pathStr := strings.Join(path, keyPathSep)
	key := m.keys[pathStr]
	if key != nil {
		return key, nil
	}
	parentKey, err := m.GetKey(path[0 : len(path)-1])
	if err != nil {
		return nil, err
	}
	childName := path[len(path)-1]
	key, err = m.deriveChildKey(parentKey, childName, pathStr)
	if err != nil {
		return nil, err
	}
	m.keys[pathStr] = key
	return key, nil
}

// Given a parentKey and a childName, derive the child's key
func (m *KeyTree) deriveChildKey(parentKey bccsp.Key, childName, path string) (bccsp.Key, error) {
	opts := &bccsp.HMACDeriveKeyOpts{
		Temporary: true,
		Arg:       []byte(childName),
	}
	key, err := m.bccspMgr.KeyDeriv(parentKey, opts)
	if err != nil {
		return nil, fmt.Errorf("Failed to derive key %s: %s", path, err)
	}
	return key, nil
}

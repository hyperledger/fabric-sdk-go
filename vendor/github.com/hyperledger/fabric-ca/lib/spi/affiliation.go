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

package spi

// AffiliationImpl defines a group name and its parent
type AffiliationImpl struct {
	Name   string `db:"name"`
	Prekey string `db:"prekey"`
}

// Affiliation is the API for a user's affiliation
type Affiliation interface {
	GetName() string
	GetPrekey() string
}

// GetName returns the name of the affiliation
func (g *AffiliationImpl) GetName() string {
	return g.Name
}

// GetPrekey returns the prekey of the affiliation
func (g *AffiliationImpl) GetPrekey() string {
	return g.Prekey
}

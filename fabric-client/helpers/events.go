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

package helpers

import "github.com/hyperledger/fabric-sdk-go/fabric-client/events"

// RegisterTxEvent registers on the given eventhub for the give transaction
// returns a boolean channel which receives true when the event is complete
// and an error channel for errors
func RegisterTxEvent(txID string, eventHub events.EventHub) (chan bool, chan error) {
	done := make(chan bool)
	fail := make(chan error)

	eventHub.RegisterTxEvent(txID, func(txId string, err error) {
		if err != nil {
			logger.Debugf("Received error event for txid(%s)\n", txId)
			fail <- err
		} else {
			logger.Debugf("Received success event for txid(%s)\n", txId)
			done <- true
		}
	})

	return done, fail
}

// RegisterCCEvent registers chain code event on the given eventhub
// @returns {chan bool} channel which receives true when the event is complete
// @returns {object} ChainCodeCBE object handle that should be used to unregister
func RegisterCCEvent(chainCodeID, eventID string, eventHub events.EventHub) (chan bool, *events.ChainCodeCBE) {
	done := make(chan bool)

	// Register callback for CE
	rce := eventHub.RegisterChaincodeEvent(chainCodeID, eventID, func(ce *events.ChaincodeEvent) {
		logger.Debugf("Received CC event: %v\n", ce)
		done <- true
	})

	return done, rce
}

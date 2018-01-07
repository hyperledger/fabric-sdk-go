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
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package consumer

import (
	"crypto/x509"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
	ehpb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

var consumerLogger = flogging.MustGetLogger("eventhub_consumer")

//EventsClient holds the stream and adapter for consumer to work with
type EventsClient struct {
	sync.RWMutex
	peerAddress string
	regTimeout  time.Duration
	stream      ehpb.Events_ChatClient
	adapter     EventAdapter
}

// RegistrationConfig holds the information to be used when registering for
// events from the eventhub
type RegistrationConfig struct {
	InterestedEvents []*ehpb.Interest
	Timestamp        *timestamp.Timestamp
	TlsCert          *x509.Certificate
}

/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gossip

import (
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
)

// ToGossipMessage un-marshals a given envelope and creates a
// SignedGossipMessage out of it.
// Returns an error if un-marshaling fails.
func (e *Envelope) ToGossipMessage() (*SignedGossipMessage, error) {
	if e == nil {
		return nil, errors.New("nil envelope")
	}
	msg := &GossipMessage{}
	err := proto.Unmarshal(e.Payload, msg)
	if err != nil {
		return nil, fmt.Errorf("Failed unmarshaling GossipMessage from envelope: %v", err)
	}
	return &SignedGossipMessage{
		GossipMessage: msg,
		Envelope:      e,
	}, nil
}

// InternalEndpoint returns the internal endpoint
// in the secret envelope, or an empty string
// if a failure occurs.
func (s *SecretEnvelope) InternalEndpoint() string {
	secret := &Secret{}
	if err := proto.Unmarshal(s.Payload, secret); err != nil {
		return ""
	}
	return secret.GetInternalEndpoint()
}

// SignedGossipMessage contains a GossipMessage
// and the Envelope from which it came from
type SignedGossipMessage struct {
	*Envelope
	*GossipMessage
}

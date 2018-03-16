/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pgresolver

import "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

// Item represents any item
type Item interface {
}

// Group contains a group of Items
type Group interface {
	// Items returns all of the items
	Items() []Item

	// Equals returns true if this Group contains the same items as the given Group
	Equals(other Group) bool

	// Reduce reduces the group (which may be a hierarchy of groups) into a simple, non-hierarchical set of groups.
	// For example, given the group, G=(A and (B or C or D))
	// then G.Reduce() = [(A and B) or (A and C) or (A and D)]
	Reduce() []Group
}

// GroupOfGroups contains a set of groups.
type GroupOfGroups interface {
	// GroupOfGroups is also a Group
	Group

	// Groups returns all of the groups in this container
	Groups() []Group

	// Nof returns a set of groups that includes all possible combinations for the given threshold.
	// For example, given the group-of-groups, G=(G1, G2, G3), where G1=(A or B), G2=(C or D), G3=(E or F),
	// then:
	// - G.Nof(1) = (G1 or G2 or G3)
	// - G.Nof(2) = ((G1 and G2) or (G1 and G3) or (G2 and G3)
	// - G.Nof(3) = (G1 and G2 and G3)
	Nof(threshold int32) (GroupOfGroups, error)
}

// PeerGroup contains a group of Peers
type PeerGroup interface {
	Group
	Peers() []fab.Peer
}

// Collapsable is implemented by any group that can collapse into a simple (non-hierarchical) Group
type Collapsable interface {
	// Collapse converts a hierarchical group into a single-level group (if possible).
	// For example, say G = (A and (B and C) and (D and E) and (F or G))
	// then G.Collapse() = (A and B and C and D and E and (F or G))
	Collapse() Group
}

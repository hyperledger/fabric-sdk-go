/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pgresolver

import (
	"fmt"
	"reflect"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

// MSPPeerRetriever is a function that retrieves peers by MSPID
type MSPPeerRetriever func(mspID string) []fab.Peer

// NewGroupOfGroups returns a new group of groups
func NewGroupOfGroups(groups []Group) GroupOfGroups {
	items := make([]Item, len(groups))
	for i, g := range groups {
		items[i] = g
	}
	return &groupsImpl{groupImpl: groupImpl{Itms: items}}
}

// NewGroup creates a new Group
func NewGroup(items []Item) Group {
	return &groupImpl{Itms: items}
}

// NewPeerGroup returns a new PeerGroup
func NewPeerGroup(peers ...fab.Peer) PeerGroup {
	return &peerGroup{peers: asPeerWrappers(peers)}
}

// NewMSPPeerGroup returns a new MSP PeerGroup
func NewMSPPeerGroup(mspID string, peerRetriever MSPPeerRetriever) PeerGroup {
	return &mspPeerGroup{
		mspID:         mspID,
		peerRetriever: peerRetriever,
	}
}

type groupImpl struct {
	Itms []Item
}

func (g *groupImpl) Items() []Item {
	return g.Itms
}

func (g *groupImpl) Reduce() []Group {
	grps := asGroupsOrPanic(g.Items())
	if len(grps) == 1 {
		return grps[0].Reduce()
	}

	// Reduce each item
	var reduced []Group
	for _, g := range grps {
		reduced = append(reduced, NewGroupOfGroups(g.Reduce()))
	}

	// Collapse each group
	var collapsed []Group
	for _, g := range and(reduced) {
		if c, ok := g.(Collapsable); ok {
			g = c.Collapse()
		}
		collapsed = append(collapsed, g)
	}

	// Get rid of duplicates
	var pruned []Group
	for _, g := range collapsed {
		if !containsGroup(pruned, g) {
			pruned = append(pruned, g)
		}
	}

	return pruned
}

func (g *groupImpl) Collapse() Group {
	var collapsable []Item
	var nonCollapsable []Item
	for _, item := range g.Items() {
		if c, ok := item.(Collapsable); ok {
			for _, ci := range c.Collapse().Items() {
				if containsItem(collapsable, ci) {
					continue
				}
				collapsable = append(collapsable, ci)
			}
		} else {
			if !containsItem(nonCollapsable, item) {
				nonCollapsable = append(nonCollapsable, item)
			}
		}
	}

	if len(collapsable) == 0 {
		return NewGroup(nonCollapsable)
	}

	cg := NewGroup(collapsable)
	if len(nonCollapsable) == 0 {
		return cg
	}

	return NewGroup(append(nonCollapsable, cg))
}

func (g *groupImpl) Equals(other Group) bool {
	if len(g.Items()) != len(other.Items()) {
		return false
	}
	for _, i1 := range g.Items() {
		if !containsItem(other.Items(), i1) {
			return false
		}
	}
	return true
}

func (g *groupImpl) String() string {
	items := g.Items()
	str := ""
	if len(items) > 1 {
		str = "["
	}
	for i, item := range items {
		str = str + fmt.Sprintf("%s", item)
		if i+1 < len(items) {
			str += ", "
		}
	}
	if len(items) > 1 {
		str += "]"
	}
	return str
}

type groupsImpl struct {
	groupImpl
}

func (g *groupsImpl) Groups() []Group {
	groups := make([]Group, len(g.Items()))
	for i, item := range g.Items() {
		if group, ok := item.(Group); ok {
			groups[i] = group
		} else {
			// This shouldn't happen since we have control over how the items are set.
			panic("unexpected: item is not a Group")
		}
	}
	return groups
}

func (g *groupsImpl) Reduce() []Group {
	var result []Group
	for _, grp := range g.Groups() {
		result = append(result, grp.Reduce()...)
	}
	return result
}

func (g *groupsImpl) Collapse() Group {
	return g
}

func (g *groupsImpl) Nof(threshold int32) (GroupOfGroups, error) {
	if int(threshold) > len(g.Items()) {
		return nil, errors.New("N is greater than length of the group")
	}
	if threshold <= 0 {
		return nil, errors.New("N must be greater than 0")
	}
	return getCombinations(g.Items(), threshold, 0)
}

func (g *groupsImpl) String() string {
	groups := g.Groups()
	str := ""
	if len(groups) > 1 {
		str = "("
	}

	for i, pg := range groups {
		str = str + fmt.Sprintf("%s", pg)
		if i+1 < len(groups) {
			str += ", "
		}
	}
	if len(groups) > 1 {
		str += ")"
	}
	return str
}

// peerWrapper wraps a Peer and implements the String() function (to help in debugging).
type peerWrapper struct {
	target fab.Peer
}

func (pw *peerWrapper) String() string {
	return pw.target.URL()
}

type peerGroup struct {
	peers []*peerWrapper
}

func (pg *peerGroup) Items() []Item {
	items := make([]Item, len(pg.peers))
	for i, peer := range pg.peers {
		items[i] = peer
	}
	return items
}

func (pg *peerGroup) Peers() []fab.Peer {
	peers := make([]fab.Peer, len(pg.peers))
	for i, pw := range pg.peers {
		peers[i] = pw.target
	}
	return peers
}

func (pg *peerGroup) Equals(other Group) bool {
	if len(pg.Items()) != len(other.Items()) {
		return false
	}
	for _, item := range pg.Items() {
		if !containsItem(other.Items(), item) {
			return false
		}
	}
	return true
}

func (pg *peerGroup) String() string {
	items := pg.Items()
	str := ""
	if len(items) > 1 {
		str = "["
	}
	for i, item := range items {
		str = str + fmt.Sprintf("%s", item)
		if i+1 < len(items) {
			str += " AND "
		}
	}
	if len(items) > 1 {
		str += "]"
	}
	return str
}

func (pg *peerGroup) Reduce() []Group {
	return []Group{pg}
}

func (pg *peerGroup) Collapse() Group {
	return NewGroup([]Item{pg})
}

type mspPeerGroup struct {
	mspID         string
	peerRetriever MSPPeerRetriever
}

func (pg *mspPeerGroup) Items() []Item {
	peers := pg.Peers()
	items := make([]Item, len(peers))
	for i, peer := range peers {
		items[i] = peer
	}
	return items
}

func (pg *mspPeerGroup) Peers() []fab.Peer {
	return pg.peerRetriever(pg.mspID)
}

func (pg *mspPeerGroup) Equals(other Group) bool {
	if otherPG, ok := other.(*mspPeerGroup); ok {
		return otherPG.GetName() == pg.GetName()
	}
	return false
}

func (pg *mspPeerGroup) Reduce() []Group {
	return []Group{pg}
}

func (pg *mspPeerGroup) Collapse() Group {
	return NewGroup([]Item{pg})
}

func (pg *mspPeerGroup) String() string {
	return pg.GetName()
}

func (pg *mspPeerGroup) GetName() string {
	return pg.mspID
}

func asPeerWrappers(peers []fab.Peer) []*peerWrapper {
	items := make([]*peerWrapper, len(peers))
	for i, peer := range peers {
		items[i] = &peerWrapper{target: peer}
	}
	return items
}

// asGroupsOrPanic converts the given array of Item into an array of Group.
// Each of the given items in the array must also be a Group or else a panic results.
func asGroupsOrPanic(items []Item) []Group {
	groups := make([]Group, len(items))
	for i, item := range items {
		if grp, ok := item.(Group); ok {
			groups[i] = grp
		} else {
			panic(fmt.Sprintf("item is not a group: %s", reflect.TypeOf(item)))
		}
	}
	return groups
}

func getCombinations(items []Item, length int32, r int) (GroupOfGroups, error) {
	if length == 1 {
		// Create an item group for each item, containing a single item
		var groups []Group
		for _, item := range items {
			groups = append(groups, NewGroup([]Item{item}))
		}
		combinations := NewGroupOfGroups(groups)

		return combinations, nil
	}

	var groups []Group
	for i := 0; i < len(items)-int(length)+1; i++ {
		leftItem := items[i]
		rightCombinations, err := getCombinations(items[i+1:], length-1, r+1)
		if err != nil {
			return nil, err
		}

		// Add the leftItem to each of the groups that came back
		for _, g := range rightCombinations.Groups() {
			var newItems []Item
			newItems = append(newItems, leftItem)
			newItems = append(newItems, g.Items()...)
			groups = append(groups, NewGroup(newItems))
		}
	}
	return NewGroupOfGroups(groups), nil
}

// and performs an 'and' operation of the given set of groups
// For example, given the set of groups, G=[(A,B),(C,D)],
// then and(G) = [(A,C),(A,D),(B,C),(B,D)]
func and(groups []Group) []Group {
	op := &andOperation{stack: &stack{}}
	op.and(groups, 0)
	return op.result
}

type andOperation struct {
	stack  *stack
	result []Group
}

func (o *andOperation) and(grps []Group, index int) {
	if index >= len(grps) {
		var items []Item
		for _, c := range o.stack.Groups() {
			groupItems := c.group.Items()
			if c.index < len(groupItems) {
				items = append(items, groupItems[c.index])
			} else {
				logger.Warnf("Expecting index to be less than %d but got %d", len(groupItems), c.index)
			}
		}
		if len(items) > 0 {
			o.result = append(o.result, NewGroup(items))
		}
	} else {
		grp := grps[index]
		items := grp.Items()
		for j := 0; j < len(items); j++ {
			o.stack.Push(grps[index], j)
			o.and(grps, index+1)
			o.stack.Pop()
		}
	}
}

type stack struct {
	items []*entry
}

func (s *stack) Push(group Group, index int) {
	s.items = append(s.items, &entry{
		group: group,
		index: index,
	})
}

func (s *stack) Pop() {
	lastIndex := len(s.items) - 1
	if lastIndex >= 0 {
		s.items = s.items[0:lastIndex]
	}
}

func (s *stack) Groups() []*entry {
	return s.items
}

type entry struct {
	group Group
	index int
}

func containsItem(items []Item, item Item) bool {
	if grp, ok := item.(Group); ok {
		for _, item2 := range items {
			if ogrp, ok2 := item2.(Group); ok2 {
				if grp.Equals(ogrp) {
					return true
				}
			}
		}
	} else {
		for _, itm := range items {
			if itm == item {
				return true
			}
		}
	}
	return false
}

func containsGroup(groups []Group, group Group) bool {
	for _, g := range groups {
		if g.Equals(group) {
			return true
		}
	}
	return false
}

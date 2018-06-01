/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package pgresolver

import (
	"testing"
)

const (
	a = "A"
	b = "B"
	c = "C"
	d = "D"
	e = "E"
	f = "F"
	h = "H"
	i = "I"
	j = "J"
	k = "K"
	l = "L"
	m = "M"
)

func TestGroupItems(t *testing.T) {
	g := g(a, b)

	items := g.Items()
	if len(items) != 2 {
		t.Fatalf("expecting %d items in group but got %d", 2, len(items))
	}
	item0 := items[0]
	if s, ok := item0.(string); ok {
		if s != a {
			t.Fatalf("expecting item[%d] to be %s but got %s", 0, a, items[0])
		}
	} else {
		t.Fatalf("expecting item[%d] to be %s but got %s", 0, a, items[0])
	}
}

func TestGroupEquals(t *testing.T) {
	g1 := g(a, b)
	g2 := g(a, b)
	g3 := g(a, b, c)
	g4 := g(a, c)

	if !g1.Equals(g1) {
		t.Fatal("expecting Equals to return true")
	}
	if !g1.Equals(g2) {
		t.Fatal("expecting Equals to return true")
	}
	if g1.Equals(g3) {
		t.Fatal("expecting Equals to return false")
	}
	if g3.Equals(g4) {
		t.Fatal("expecting Equals to return false")
	}
	if !g3.Equals(g3) {
		t.Fatal("expecting Equals to return true")
	}
}

func TestGroupReduce(t *testing.T) {
	g1 := mg(a, b)
	g2 := mg(c, d)

	r1 := g(g1).Reduce()
	logger.Debugf("%v\n", r1)
	verifyGroups(t, []Group{g1}, r1)

	r2 := g(g1, g2).Reduce()
	logger.Debugf("%v\n", r2)
	verifyGroups(t, []Group{g(g1, g2)}, r2)
}

func TestGroupCollapse(t *testing.T) {
	g1 := g(g(a, b), g(c), g(d, e, f))
	logger.Debugf("%v\n", g1)

	r1 := g1.(Collapsable).Collapse()
	logger.Debugf("%v\n", r1)
	expected := g(a, b, c, d, e, f)
	if !expected.Equals(r1) {
		t.Fatalf("group %s is not in the set of expected groups: %v", g1, expected)
	}

}

func TestGOGGroups(t *testing.T) {
	g1 := mg(a, b)
	g2 := mg(c, d)

	gog1 := gog(g1, g2)

	groups := gog1.Groups()
	if len(groups) != 2 {
		t.Fatalf("expecting %d groups in the group-of-groups but got %d", 2, len(groups))
	}
	group0 := groups[0]
	if g, ok := group0.(Group); ok {
		if g != g1 {
			t.Fatalf("expecting item[%d] to be %s but got %s", 0, g1, groups[0])
		}
	} else {
		t.Fatalf("expecting item[%d] to be %s but got %s", 0, g1, groups[0])
	}
}

func TestGOGEquals(t *testing.T) {
	g1 := g(a, b)
	g2 := g(a, b)
	g3 := g(a, b, c)
	g4 := g(a, c)

	gog1 := gog(g1, g2)
	gog2 := gog(g1, g2)
	gog3 := gog(g3, g4)
	gog4 := gog(g3, g4)

	if !gog1.Equals(gog1) {
		t.Fatal("expecting Equals to return true")
	}
	if !gog1.Equals(gog2) {
		t.Fatal("expecting Equals to return true")
	}
	if gog1.Equals(gog3) {
		t.Fatal("expecting Equals to return false")
	}
	if !gog3.Equals(gog4) {
		t.Fatal("expecting Equals to return true")
	}
}

func TestGroupOfGroupsReduce(t *testing.T) {
	g1 := mg(a, b)
	g2 := mg(c, d)

	r1 := gog(g1).Reduce()
	logger.Debugf("%v\n", r1)
	verifyGroups(t, []Group{g1}, r1)

	r2 := gog(g1, g2).Reduce()
	logger.Debugf("%v\n", r2)
	verifyGroups(t, []Group{g1, g2}, r2)
}

func TestGOGCollapse(t *testing.T) {
	g1 := gog(g(a, b), g(c), g(d, e, f))
	logger.Debugf("%v\n", g1)

	r1 := g1.(Collapsable).Collapse()
	logger.Debugf("%v\n", r1)
	expected := g1
	if !expected.Equals(r1) {
		t.Fatalf("group %s is not in the set of expected groups: %v", g1, expected)
	}
}

func TestCompositeReduce(t *testing.T) {
	g1 := mg(a, b)
	g2 := mg(c, d)
	g3 := mg(e, f)
	g4 := mg(h, i, j)
	g5 := mg(k, l, m)

	r := g(gog(g1, g2), gog(g3, g4, g5)).Reduce()
	logger.Debugf("%v\n", r)
	verifyGroups(t, []Group{g(g1, g3), g(g1, g4), g(g1, g5), g(g2, g3), g(g2, g4), g(g2, g5)}, r)
}

func TestAndOperation(t *testing.T) {
	g1 := g(a, b)
	g2 := g(c, d, e)

	expected := []Group{
		g(a, c), g(a, d), g(a, e),
		g(b, c), g(b, d), g(b, e),
	}

	r := and([]Group{g1, g2})
	logger.Debugf("%v\n", r)

	verifyGroups(t, expected, r)
}

func g(items ...Item) Group {
	return NewGroup(items)
}

func gog(groups ...Group) GroupOfGroups {
	return NewGroupOfGroups(groups)
}

func verifyGroups(t *testing.T, expected []Group, actual []Group) {
	if len(expected) != len(actual) {
		t.Fatalf("expecting %d groups but got %d", len(expected), len(actual))
	}

	for _, g := range actual {
		if !containsGroup(expected, g) {
			t.Fatalf("group %s is not in the set of expected groups: %v", g, expected)
		}
	}
}

type mockGroup struct {
	groupImpl
}

func (g *mockGroup) Reduce() []Group {
	return []Group{g}
}

func (g *mockGroup) Collapse() Group {
	return NewGroup([]Item{g})
}

func mg(items ...Item) Group {
	itms := make([]Item, len(items))
	for i := 0; i < len(items); i++ {
		itms[i] = items[i]
	}
	return &mockGroup{groupImpl{Itms: itms}}
}

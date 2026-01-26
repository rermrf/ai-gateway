package loadbalancer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockNode struct {
	id string
}

func (n *mockNode) ID() string {
	return n.id
}

func TestRoundRobin(t *testing.T) {
	nodes := []Node{&mockNode{"1"}, &mockNode{"2"}, &mockNode{"3"}}
	lb := NewRoundRobin[Node](nodes)

	// Should cycle through 1 -> 2 -> 3 -> 1
	n1, _ := lb.Select()
	assert.Equal(t, "1", n1.ID())

	n2, _ := lb.Select()
	assert.Equal(t, "2", n2.ID())

	n3, _ := lb.Select()
	assert.Equal(t, "3", n3.ID())

	n4, _ := lb.Select()
	assert.Equal(t, "1", n4.ID())
}

func TestRoundRobin_Empty(t *testing.T) {
	lb := NewRoundRobin[Node](nil)
	n, err := lb.Select()
	assert.Error(t, err)
	assert.Nil(t, n)
}

func TestRandom(t *testing.T) {
	nodes := []Node{&mockNode{"1"}, &mockNode{"2"}}
	lb := NewRandom[Node](nodes)

	// Just checking it returns a valid node
	n, err := lb.Select()
	assert.NoError(t, err)
	assert.Contains(t, []string{"1", "2"}, n.ID())
}

func TestWeighted(t *testing.T) {
	nodes := []Node{&mockNode{"A"}, &mockNode{"B"}}
	weights := []int{1, 0} // B should never be selected
	lb := NewWeighted[Node](nodes, weights)

	for i := 0; i < 10; i++ {
		n, err := lb.Select()
		assert.NoError(t, err)
		assert.Equal(t, "A", n.ID())
	}
}

func TestFailover(t *testing.T) {
	nodes := []Node{&mockNode{"Primary"}, &mockNode{"Backup"}}
	lb := NewFailover[Node](nodes)

	// Should always return Primary
	n1, _ := lb.Select()
	assert.Equal(t, "Primary", n1.ID())

	n2, _ := lb.Select()
	assert.Equal(t, "Primary", n2.ID())
}

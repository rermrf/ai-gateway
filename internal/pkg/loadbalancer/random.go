// Package loadbalancer 提供通用的负载均衡算法。
package loadbalancer

import (
	"math/rand"
)

// Random 实现随机负载均衡。
type Random[T Node] struct {
	nodes []T
}

// NewRandom 创建一个新的随机负载均衡器。
func NewRandom[T Node](nodes []T) *Random[T] {
	return &Random[T]{
		nodes: nodes,
	}
}

func (r *Random[T]) Select() (T, error) {
	var zero T
	if len(r.nodes) == 0 {
		return zero, ErrNoAvailableNode
	}
	return r.nodes[rand.Intn(len(r.nodes))], nil
}

func (r *Random[T]) ReportSuccess(node T) {}
func (r *Random[T]) ReportFailure(node T) {}

func (r *Random[T]) Nodes() []T {
	return r.nodes
}

func (r *Random[T]) UpdateNodes(nodes []T) {
	r.nodes = nodes
}

var _ LoadBalancer[Node] = (*Random[Node])(nil)

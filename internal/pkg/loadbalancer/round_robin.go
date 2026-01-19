// Package loadbalancer 提供通用的负载均衡算法。
package loadbalancer

import (
	"sync/atomic"
)

// RoundRobin 实现轮询负载均衡。
type RoundRobin[T Node] struct {
	nodes   []T
	counter uint64
}

// NewRoundRobin 创建一个新的轮询负载均衡器。
func NewRoundRobin[T Node](nodes []T) *RoundRobin[T] {
	return &RoundRobin[T]{
		nodes: nodes,
	}
}

func (r *RoundRobin[T]) Select() (T, error) {
	var zero T
	if len(r.nodes) == 0 {
		return zero, ErrNoAvailableNode
	}
	idx := atomic.AddUint64(&r.counter, 1) - 1
	return r.nodes[idx%uint64(len(r.nodes))], nil
}

func (r *RoundRobin[T]) ReportSuccess(node T) {}
func (r *RoundRobin[T]) ReportFailure(node T) {}

func (r *RoundRobin[T]) Nodes() []T {
	return r.nodes
}

func (r *RoundRobin[T]) UpdateNodes(nodes []T) {
	r.nodes = nodes
}

var _ LoadBalancer[Node] = (*RoundRobin[Node])(nil)

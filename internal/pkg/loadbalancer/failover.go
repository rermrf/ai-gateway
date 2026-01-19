// Package loadbalancer 提供通用的负载均衡算法。
package loadbalancer

import (
	"sync"
)

// Failover 实现故障转移负载均衡。
// 它按顺序尝试节点，跳过失败的节点，直到报告成功。
type Failover[T Node] struct {
	nodes  []T
	failed map[string]bool
	mu     sync.RWMutex
}

// NewFailover 创建一个新的故障转移负载均衡器。
// 节点按优先级排序（第一个为主要节点）。
func NewFailover[T Node](nodes []T) *Failover[T] {
	return &Failover[T]{
		nodes:  nodes,
		failed: make(map[string]bool),
	}
}

func (f *Failover[T]) Select() (T, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var zero T
	for _, node := range f.nodes {
		if !f.failed[node.ID()] {
			return node, nil
		}
	}
	// 全部失败，仍然尝试第一个节点
	if len(f.nodes) > 0 {
		return f.nodes[0], nil
	}
	return zero, ErrNoAvailableNode
}

func (f *Failover[T]) ReportSuccess(node T) {
	f.mu.Lock()
	f.failed[node.ID()] = false
	f.mu.Unlock()
}

func (f *Failover[T]) ReportFailure(node T) {
	f.mu.Lock()
	f.failed[node.ID()] = true
	f.mu.Unlock()
}

func (f *Failover[T]) Nodes() []T {
	return f.nodes
}

func (f *Failover[T]) UpdateNodes(nodes []T) {
	f.mu.Lock()
	f.nodes = nodes
	f.mu.Unlock()
}

var _ LoadBalancer[Node] = (*Failover[Node])(nil)

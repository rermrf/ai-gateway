// Package loadbalancer 提供通用的负载均衡算法。
package loadbalancer

import (
	"math/rand"
)

// Weighted 实现加权随机负载均衡。
type Weighted[T Node] struct {
	nodes       []T
	weights     []int
	totalWeight int
}

// NewWeighted 创建一个新的加权负载均衡器。
// weights 切片的长度必须与 nodes 相同。
func NewWeighted[T Node](nodes []T, weights []int) *Weighted[T] {
	total := 0
	for _, w := range weights {
		total += w
	}
	return &Weighted[T]{
		nodes:       nodes,
		weights:     weights,
		totalWeight: total,
	}
}

// NewWeightedFromNodes 从 WeightedNode 切片创建一个加权负载均衡器。
func NewWeightedFromNodes[T Node](weightedNodes []WeightedNode[T]) *Weighted[T] {
	nodes := make([]T, len(weightedNodes))
	weights := make([]int, len(weightedNodes))
	for i, wn := range weightedNodes {
		nodes[i] = wn.Node
		weights[i] = wn.Weight
	}
	return NewWeighted(nodes, weights)
}

func (w *Weighted[T]) Select() (T, error) {
	var zero T
	if len(w.nodes) == 0 || w.totalWeight == 0 {
		return zero, ErrNoAvailableNode
	}

	r := rand.Intn(w.totalWeight)
	for i, weight := range w.weights {
		r -= weight
		if r < 0 {
			return w.nodes[i], nil
		}
	}
	return w.nodes[0], nil
}

func (w *Weighted[T]) ReportSuccess(node T) {}
func (w *Weighted[T]) ReportFailure(node T) {}

func (w *Weighted[T]) Nodes() []T {
	return w.nodes
}

func (w *Weighted[T]) UpdateNodes(nodes []T) {
	w.nodes = nodes
	// 注意：权重也应相应更新
}

// UpdateNodesWithWeights 同时更新节点和权重。
func (w *Weighted[T]) UpdateNodesWithWeights(nodes []T, weights []int) {
	w.nodes = nodes
	w.weights = weights
	total := 0
	for _, weight := range weights {
		total += weight
	}
	w.totalWeight = total
}

var _ LoadBalancer[Node] = (*Weighted[Node])(nil)

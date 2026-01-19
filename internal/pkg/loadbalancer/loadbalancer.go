// Package loadbalancer 提供通用的负载均衡算法。
// 它使用 Go 泛型来处理任何实现 Node 接口的类型。
package loadbalancer

import (
	"errors"
)

// ErrNoAvailableNode 当没有可用节点时返回。
var ErrNoAvailableNode = errors.New("no available node")

// Node 表示可以进行负载均衡的节点。
// 任何可以提供唯一标识符的类型都可以作为 Node。
type Node interface {
	// ID 返回此节点的唯一标识符。
	ID() string
}

// LoadBalancer 是通用的负载均衡器接口。
type LoadBalancer[T Node] interface {
	// Select 返回下一个要使用的节点。
	Select() (T, error)
	// ReportSuccess 报告节点的请求成功。
	ReportSuccess(node T)
	// ReportFailure 报告节点的请求失败。
	ReportFailure(node T)
	// Nodes 返回负载均衡器中的所有节点。
	Nodes() []T
	// UpdateNodes 更新节点列表。
	UpdateNodes(nodes []T)
}

// WeightedNode 包装一个带有权重的节点。
type WeightedNode[T Node] struct {
	Node   T
	Weight int
}

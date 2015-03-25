package strategy

import (
	"github.com/docker/swarm/cluster"
	"github.com/samalba/dockerclient"
)

// WeightedNode represents a node in the cluster with a given weight, typically used for sorting
// purposes.
type weightedNode struct {
	Node cluster.Node
	// Weight is the inherent value of this node.
	Weight float64
}

type weightedNodeList []*weightedNode

func (n weightedNodeList) Len() int {
	return len(n)
}

func (n weightedNodeList) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n weightedNodeList) Less(i, j int) bool {
	var (
		ip = n[i]
		jp = n[j]
	)

	return ip.Weight < jp.Weight
}

func weighNodes(config *dockerclient.ContainerConfig, nodes []cluster.Node) (weightedNodeList, error) {
	weightedNodes := weightedNodeList{}

	for _, node := range nodes {
		nodeMemory := node.TotalMemory()
		nodeCpus := node.TotalCpus()

		// Skip nodes that are smaller than the requested resources.
		if nodeMemory < int64(config.Memory) || nodeCpus < normalize(config.CpuShares, node.GetCpus()) {
			continue
		}

		var (
			cpuScore    float64 = 100
			memoryScore int64   = 100
		)

		if config.CpuShares > 0 {
			cpuScore = (node.UsedCpus() + normalize(config.CpuShares, node.GetCpus())) * 100 / nodeCpus
		}
		if config.Memory > 0 {
			memoryScore = (node.UsedMemory() + config.Memory) * 100 / nodeMemory
		}

		if cpuScore <= 100 && memoryScore <= 100 {
			weightedNodes = append(weightedNodes, &weightedNode{Node: node, Weight: cpuScore + float64(memoryScore)})
		}
	}

	if len(weightedNodes) == 0 {
		return nil, ErrNoResourcesAvailable
	}

	return weightedNodes, nil
}

func normalize(cpushare int64, cpus int64) float64 {
	return float64(cpushare*cpus) / 1024
}

package agent

import (
	"github.com/eirsyl/apollo/pkg"
	"github.com/eirsyl/apollo/pkg/agent/redis"
	pb "github.com/eirsyl/apollo/pkg/api"
)

/**
 * This file contains internal helper functions used by the agent only
 */

func transformNodes(nodes *[]redis.ClusterNode) *[]*pb.StateRequest_ClusterNode {
	var tNodes = make([]*pb.StateRequest_ClusterNode, len(*nodes))

	for i, node := range *nodes {
		tNode := pb.StateRequest_ClusterNode{
			NodeID:      node.NodeID,
			Addr:        node.Addr,
			Flags:       node.Flags,
			Role:        node.Role,
			Myself:      node.Myself,
			MasterID:    node.MasterID,
			PingSent:    int64(node.PingSent),
			PingRecv:    int64(node.PingRecv),
			ConfigEpoch: int64(node.ConfigEpoch),
			LinkStatus:  node.LinkStatus,
			Slots:       node.Slots,
		}
		tNodes[i] = &tNode
	}

	return &tNodes
}

func transformHostAnnotations(hostAnnotations *map[string]string) *[]*pb.StateRequest_HostAnnotation {
	var tAnnotations = []*pb.StateRequest_HostAnnotation{}

	for name, value := range *hostAnnotations {
		tAnnotation := pb.StateRequest_HostAnnotation{
			Name:  name,
			Value: value,
		}

		tAnnotations = append(tAnnotations, &tAnnotation)
	}

	return &tAnnotations
}

func transformMetrics(metrics *chan Metric) *[]*pb.StateRequest_NodeMetric {
	var result []*pb.StateRequest_NodeMetric

	for metric := range *metrics {
		if pkg.AgentMetrics[metric.Name] {
			result = append(result, &pb.StateRequest_NodeMetric{Name: metric.Name, Value: metric.Value})
		}
	}

	return &result
}

package agent

import (
	"github.com/eirsyl/apollo/pkg/agent/redis"
	pb "github.com/eirsyl/apollo/pkg/api"
)

func transformNodes(nodes *[]redis.ClusterNode) *[]*pb.ClusterNode {
	var tNodes = make([]*pb.ClusterNode, len(*nodes))

	for i, node := range *nodes {
		tNode := pb.ClusterNode{
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

func transformHostAnnotations(hostAnnotations *map[string]string) *[]*pb.HostAnnotation {
	var tAnnotations = []*pb.HostAnnotation{}

	for name, value := range *hostAnnotations {
		tAnnotation := pb.HostAnnotation{
			Name:  name,
			Value: value,
		}

		tAnnotations = append(tAnnotations, &tAnnotation)
	}

	return &tAnnotations
}

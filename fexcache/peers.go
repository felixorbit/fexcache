package fexcache

import (
	pb "felixorb/fexcache/fexcachepb"
)

type PeerPicker interface {
	PickPeer(string) (PeerGetter, bool)
}

// 分布式缓存的目的是不同key缓存在不同的节点上，增加总的吞吐量
type PeerGetter interface {
	Get(*pb.Request, *pb.Response) error
}

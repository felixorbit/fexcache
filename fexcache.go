package fexcache

import (
	"fmt"
	"log"
	"sync"

	pb "github.com/felixorbit/fexcache/fexcachepb"
	"github.com/felixorbit/fexcache/singleflight"
)

// Group 是核心结构。负责与外部交互
type Group struct {
	name      string // 命名空间
	mainCache cache
	peers     PeerPicker // 远程节点选择器
	getter    Getter     // 缓存不存在时的回调接口
	loader    *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// Getter 是缓存不存在时的回调接口。函数类型或结构体类型都可以实现该接口
type Getter interface {
	Get(string) ([]byte, error)
}

// GetterFunc 是接口型函数
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) { return f(key) }

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	g := &Group{
		name:      name,
		mainCache: cache{cacheBytes: cacheBytes},
		getter:    getter,
		loader:    &singleflight.Group{},
	}
	mu.Lock()
	defer mu.Unlock()
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// 调用回调函数获取数据，添加到缓存
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: bytes}
	g.populateCache(key, value)
	return value, nil
}

// 从其他远程节点获取数据
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{Group: g.name, Key: key}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

// 实现 singleflight
func (g *Group) load(key string) (value ByteView, err error) {
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if view, err := g.getFromPeer(peer, key); err == nil {
					return view, nil
				}
				log.Println("Failed to get from peer")
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

// Get 获取数据的主要逻辑
// 1. 本节点查询
// 2. 远程节点查询
// 3. 缓存不存在，执行回调函数
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		return v, nil
	}
	return g.load(key)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeers called more than once")
	}
	g.peers = peers
}

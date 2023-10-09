package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

/**
一致性哈希算法将 key 映射到 2^32 的空间中，将这个数字首尾相连，形成一个环
计算节点/机器(通常使用节点的名称、编号和 IP 地址)的哈希值，放置在环上。
计算 key 的哈希值，放置在环上，顺时针寻找到的第一个节点，就是应选取的节点/机器

（两个不同的集合，通过哈希函数映射到同一个（一维）空间中，按照一定的方向进行匹配，从而得到两个集合的一一映射关系）
*/

type Hash func(data []byte) uint32

// Map 实现一致性哈希
type Map struct {
	hash     Hash
	keys     []int          // 保存所有节点，需要维持有序，方便查找
	replicas int            // 添加虚拟节点解决数据倾斜
	hashMap  map[int]string // 保存虚拟节点-真实节点的映射
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hashK := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hashK)
			m.hashMap[hashK] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hashK := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hashK
	})
	// 找不到第一个大于 key 的节点时会返回 len(m.keys)，需要取余实现环形
	return m.hashMap[m.keys[idx%len(m.keys)]]
}

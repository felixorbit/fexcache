package fexcache

type PeerPicker interface {
	PickPeer(string) (PeerGetter, bool)
}

type PeerGetter interface {
	Get(group, key string) ([]byte, error)
}

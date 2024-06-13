package fexcache

type ByteView struct {
	b []byte
}

func (v ByteView) Len() int { return len(v.b) }

func cloneBytes(bytes []byte) []byte {
	newBytes := make([]byte, len(bytes))
	copy(newBytes, bytes)
	return newBytes
}

func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

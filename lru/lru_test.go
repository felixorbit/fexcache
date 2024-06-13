package lru

import (
	"reflect"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestGet(t *testing.T) {
	lru := NewCache(int64(0), nil)
	lru.Add("foo", String("bar"))
	if v, ok := lru.Get("foo"); !ok || string(v.(String)) != "bar" {
		t.Fatalf("cache get foo failed")
	}
	if _, ok := lru.Get("foo1"); ok {
		t.Fatalf("cache miss foo1 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	lru := NewCache(int64(11), nil)
	lru.Add("foo", String("bar"))
	lru.Add("fex", String("new"))
	if _, ok := lru.Get("foo"); ok {
		t.Fatalf("cache remove oldest failed")
	}
}

func TestOnEvicted(t *testing.T) {
	removedKeys := make([]string, 0)
	lru := NewCache(8, func(s string, v Value) {
		removedKeys = append(removedKeys, s)
	})
	lru.Add("k1", String("11"))
	lru.Add("k2", String("22"))
	lru.Add("k3", String("33"))
	lru.Add("k4", String("44"))
	expected := []string{"k1", "k2"}
	if !reflect.DeepEqual(expected, removedKeys) {
		t.Fatalf("Call OnEvicted failed, expected keys: %v, got: %v", expected, removedKeys)
	}
}

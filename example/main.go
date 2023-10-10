package main

import (
	"felixorb/fexcache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

// HTTPPool 一方面作为服务端响应远程的请求，一方面作为客户端，请求其他远程节点
// 1. HTTPPool 作为服务端，【依赖】 Group 完成获取缓存的主要逻辑，
// 2. Group 在执行时需要【关联】HTTPPool 与其他远程节点通信
// 如果将 HTTPPool 的功能拆分出服务端和客户端，关系应该更清晰一些

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *fexcache.Group {
	return fexcache.NewGroup("scores", 2<<10, fexcache.GetterFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key: ", key)
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not found", key)
	}))
}

// CacheServer 用于节点间通信
func startCacheServer(addr string, addrs []string, group *fexcache.Group) {
	pool := fexcache.NewHTTPPool(addr)
	pool.SetPeers(addrs...)
	group.RegisterPeers(pool)
	log.Println("fexcache is running at: ", addr)
	log.Fatal(http.ListenAndServe(addr[7:], pool))
}

// APIServer 与用户交互
func startAPIServer(apiAddr string, group *fexcache.Group) {
	http.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		view, err := group.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(view.ByteSlice())
	}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	group := createGroup()
	if api {
		go startAPIServer(apiAddr, group)
	}
	startCacheServer(addrMap[port], addrs, group)
}

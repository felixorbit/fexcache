# 分布式缓存 - FexCache

> 参考 [7 days golang programs from scratch](https://github.com/geektutu/7days-golang)

## 特性

- 缓存淘汰策略：LRU
- 并发访问：互斥锁
- 分布式缓存：基于 HTTP 协议的客户端/服务端通信；基于一致性哈希算法选择节点
- 防止缓存击穿：实现 Singleflight 合并并发请求
- Probobuf 序列化：（可有可无的功能）

```
                            是
接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
                |  否                         是
                |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
                            |  否
                            |-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶
```


## 安装 Protobuf

On MacOS

```
brew install protobuf

go install github.com/golang/protobuf/protoc-gen-go
```

将 `$GOPATH/bin` 加入环境变量中
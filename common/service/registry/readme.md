# 实施方案

microregistry.go 是registry的内存实现
spegel.go 是registry的etcd实现
这两都是http.Handler 类型
写一个registry.go
优先经过microregistry.go 处理, 如果不是200 则通过spegel.go 处理 

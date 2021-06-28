# 进展
## project1
* engine_util模块封装了关于badger的所有操作以支持CF，实验一就是调用engine_util的接口实现standalone_storage。
* 可以参考raft_storage来写。
* Reader()的返回值是一个storage.StorageReader接口，这个接口需要自定义结构体进行实现。
* raft_storage和standalone_storage是两个独立的storage。

- [ ] `standalone_storage.go`
- [ ] `server.go`
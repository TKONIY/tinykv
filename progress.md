# 进展
## project1
* engine_util模块封装了关于badger的所有操作以支持CF，实验一就是调用engine_util的接口实现standalone_storage。
* 可以参考raft_storage来写。
* Reader()的返回值是一个storage.StorageReader接口，这个接口需要自定义结构体进行实现。
* raft_storage和standalone_storage是两个独立的storage
- [*] `standalone_storage.go`
- [*] `server.go`

## project2
* AppendEntries和HeartBeat分成了两种消息
* 查阅doc.go
### RAFT
* commitIndex是不是就是写入到我这里来的最新的?lastLogIndex是不是就是这个
* 为什么要先判断term再判断index，不能直接判断index吗? 
* 哪个东西记录了log的长度?是commitIndex吗?判断用的index是最后的index吗
* 如果收到的vote都得到了回复，但是都是false，那应该怎么做，new election吗？
* 
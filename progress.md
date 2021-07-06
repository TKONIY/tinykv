# 进展
## project1
* engine_util模块封装了关于badger的所有操作以支持CF，实验一就是调用engine_util的接口实现standalone_storage。
* 可以参考raft_storage来写。
* Reader()的返回值是一个storage.StorageReader接口，这个接口需要自定义结构体进行实现。
* raft_storage和standalone_storage是两个独立的storage
- [x] `standalone_storage.go`
- [x] `server.go`

## project2

* AppendEntries和HeartBeat分成了两种消息
* 查阅doc.go
### RAFT
* commitIndex是不是就是写入到我这里来的最新的?lastLogIndex是不是就是这个
* 为什么要先判断term再判断index，不能直接判断index吗? 
* 哪个东西记录了log的长度?是commitIndex吗?判断用的index是最后的index吗
* 如果收到的vote都得到了回复，但是都是false，那应该怎么做，new election吗？

### `AppendEntries` RPC

#### Leader

```python
nextIndex  = [] 	# 记录了已发送给每个follower的最新log
matchIndex = []		# 记录了每个follower确认的最新log ???
```

```python
for i, f in enumerate(followers): 		# 遍历所有follower
		if log[-1].index > nextIndex[i]: 	# lastLogIndex: log[-1].index
      
        while True:										# 不断尝试直到同步成功
          
            rpc = AppendEntries();

            rpc.term 					= currentTerm;
            rpc.prevLogIndex 	= nextIndex[i] - 1;
            rpc.precLogTerm  	= log[nextIndex[i] - 1].term;
            rpc.entries				= log[nextIndex[i] : -1];
            rpc.leaderCommit  = commitIndex;

            reply = f.CallRPC(rpc);

            if reply.success == True:
                nextIndex[i]  = log[-1].index + 1;
                matchIndex[i] = log[-1].index;
            else: 											# reply.success == False
                if reply.term > currTerm:
                    SwitchMyRoleTo(FOLLOWER);
                else:										# prevLogIndex and prevLogTerm not match
                    nextIndex[i] -= 1;
                    continue;						

            break;
            
            
newCommitIndex = commitIndex;
while True:
  	newCommitIndex += 1;
    
    if newCommitIndex > len(log):
      	newCommitIndex -= 1;
      	break;
  	if log[newCommitIndex].term != currentTerm:
      	continue;
        
		count = 0
    for i, _ in enumerate(followers):
				if matchIndex[i] >= newCommitIndex:
          	count += 1;
    if count < len(followers)//2+1:
      	newCommitIndex -= 1;
        break;
    # else: count > half of numberss: continue;
commitIndex = newCommitIndex;
```

### Follower

```python
def handleAppendEntries(rpc):
  
  	reply = AppendEntriesReply();
    
    # 1.
    reply.term = currTerm;
  	if rpc.term < currTerm:
				reply.success = False;
        return reply;
      
    # 2. 
    if log[rpc.preLogIndex].term != rpc.preLogTerm:
      	reply.success = False;
        return reply;
      
    # 3&4.
    for i, entry in enumerate(rpc.entries):
      	index = i + 1 + rpc.preLogIndex;
        log[index] = entry; # 直接覆盖掉所有的, 此处忽略扩容的处理
        
    # 5.
    if rpc.leaderCommit > commitIndex:
      	commitIndex = min(rpc.leaderCommit, log[-1].index);
    
		reply.success = True;
    return reply;
```


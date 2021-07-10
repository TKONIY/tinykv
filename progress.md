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

* 没有leader的时候请求应该重定向给谁？

* 不同节点上，索引相同，日志不一定相同。但是如果同时term也相同，则日志一定相同。

    * 可能出现不同任期的leader往同一个位置写的情况，比如旧的leader不知道选举出了新的leader，所以往7里面写了一条日志，新的leader从6开始更新，也往7里面写了一条日志。
    * 但是同一个term里面，只有一个leader，而一个leader只会增量地往log里面写入，不会重复往一个index里面写入，因而可以保证**同一个term里面出现在同一个位置的日志一定相同**。

* 数学归纳法的魅力：

* 每次选举完成后，必须要先同步一次空操作才能接受客户端写请求，根本原因是必须做到日志上的同步，机器重启后并且立刻当选为leader的时候可能不知道自己之前的到底哪些commit了，**它复活了之后会以为上一个term的自己的也被commit了，实际上可能还没同步完呢**

    *   加上一个op空操作，保证了只要我当了领导，一定有大部分人知道这件事情。
    *   不会我当了领导但是还有一个迷糊蛋，它傻乎乎地当选了还覆盖了我的log。

    ![image-20210708010105384](/Users/dengyangshen.ys/Library/Application Support/typora-user-images/image-20210708010105384.png)

* 定时器要求：广播时延<<心跳超时时间 << 平均故障时间

    * 时延如果比心跳超时要长那就一直超时了心跳没用
    * 故障如果比心跳更频繁那心跳检测不到故障了

    >   客户端和服务器之间要保证不要重复（raft上层的内容)
    >
    >   保证读写一致性：每次写入/读取成功后返回给客户端lastapplied，如果有follower接收到读请求则先检查自己的数据是不是最新的。
    >
    >   Split-brain的时候，老的leader还是可以读写：写不会成功因为无法同步，读会成功
-   [ ] 是否需要实现线性读，如何实现：看看pingcap的文档。是只读leader吗？

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

*   其他问题：
    *   线性读如何实现？
    *   Split brain，读错了leader怎么办。
*   避免太长时间的不可用时间差
    *   网络分区的时候，小数分区的节点会一直变成candidate，一直+1，最后他们的term会比大数分区的高很多。
    *   重新连接后，小数分区的节点会发起一轮选举，并且会导致原来的leader变成follower。虽然一致性得到保证，但是多了很多次选举。即使这次没选成，下一次大分区的节点重新成为了leader，但是小数分区的节点term远远大于大分区，所以它没法被同步，只是会从candidate变成follower，一直不可用。等到下一次它变成candidate时，又出来捣蛋了。
    *   解决方法是增加pre-vote RPC，加一个心跳，只有确保它在大分区里面才变成candidate。

*   集群变更
    *   先不看
*   快照
    *   先不看
*   https://github.com/etcd-io/etcd/blob/main/raft/raft.go 关于step的实现。
*   Step状态机是如何实现的。
*   ready结构体用于状态机变更后向应用层返回结果(进行应用层操作如网络操作)
*   全异步：node --channel--> raft状态机 --channel--> node
    *   保证顺序即可

### 消息处理进度

-   [ ] MsgHup
    *   tickElection() -> step(msg)
    *   很长时间没收到heartbeat
-   [ ] MsgBeat
    *   tickHeartbeat()-> send(msg)
    *   发送心跳
-   [ ] MsgPropose
    *   不同人收到的做法不一样
    *   leader调用bcast广播
-   [ ] MsgAppend
    -   [ ] bcast的广播
-   [ ] MsgAppendResponse
    -   [ ] handleAppendEntries() -> MsgAppendResponse()
-   [ ] MsgRequestVote
    -   [ ] compain()里面调用发送
-   [ ] MsgRequestVoteResponse
    -   [ ] ...
-   [ ] MsgSnapshot
    -   [ ] 
-   [ ] MsgHeartbeat
-   [ ] MsgHeartbeatResponse
-   [ ] MsgTransferLeader
-   [ ] MsgTimeoutNow

```go
Raft {
  id,
  Term, 		// currentTerm
  Vote, 		// votefor
  RaftLog, 	// log's attr
  Prs,			// id -> MatchIndex, NextIndex
  
  State,
  msgs,
  Lead,			// leader's id
  
  heartbeatTimeout,
  electionTimeout,
  
  heartbeatElapsed,
  electionElapsed,
  
  leadTransferee,
  PendingConfIndex,
}
```

```go
RaftLog {
  storage,
  
  committed, 	// commitIndex
  applied,		// lastApplied
  stabled,			 
  entries,
  pendingSnapshot,
}
/*
        stabled applied committed
            |     |       |
+-----------*-----*-------+
| persisted |     |       |
+-----------+-----+-------+
*/
```


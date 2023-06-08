# consensusLockz

## Consus

The reasons for using `Consus` to implement distributed locks

|     | Item | Redis                                | Consus                                                      |
| :--: | :------- | :------------------------------------ | :----------------------------------------------------------- |
| C  | Consistency | Use `RDB files` snapshots to maintain consistency | Fully implement `the Raft algorithm`                      |
| A  | High availability | Excellent performance, `a single node` can withstand high concurrency | The performance is not as good as Redis, so Goroutines must `be distributed on different nodes` to read data |
| P  | Partition tolerance  | Can only be `a single node`, single point of failure,no partition tolerance | Some nodes fail and still operate |

As long as it can be used for `service discovery`, `CP` is good enough. 

# RAFT
A Golang implementation of the Raft distributed consensus protocol.

## Features

* Leader election
* Log replication
* Log compaction/Snapshot
* [RPC](https://hslam.com/git/x/rpc "rpc") transport
* ReadIndex/Lease

## [Example raftdb](https://hslam.com/git/x/raft/src/master/example/raftdb "raftdb") Benchmark

#### Linux Environment
* **CPU** 12 Cores 3.1 GHz
* **Memory** 24 GiB

```
cluster     operation   transport   requests/s  average(ms) fastest(ms) median(ms)  p99(ms)     slowest(ms)
Singleton   ReadIndex   HTTP        74456       6.63        2.62        6.23        12.12       110.90
Singleton   ReadIndex   RPC         293865      13.14       4.09        12.14       31.22       35.09
Singleton   Write       HTTP        57488       8.79        2.19        7.68        24.00       119.71
Singleton   Write       RPC         132045      30.21       6.39        27.59       70.14       86.11
ThreeNodes  ReadIndex   HTTP        43053       11.72       2.83        7.43        58.92       1125.58
ThreeNodes  ReadIndex   RPC         267685      14.65       4.36        13.47       31.59       44.72
ThreeNodes  Write       HTTP        35241       14.42       4.21        10.41       73.28       114.84
ThreeNodes  Write       RPC         103035      38.82       8.91        38.90       76.74       88.05
```

#### HTTP WRITE THREE NODES BENCHMARK
```
Summary:
	Clients:	512
	Parallel calls per client:	1
	Total calls:	100000
	Total time:	2.84s
	Requests per second:	35241.64
	Fastest time for request:	4.21ms
	Average time per request:	14.42ms
	Slowest time for request:	114.84ms

Time:
	10%	time for request:	7.84ms
	50%	time for request:	10.41ms
	90%	time for request:	24.42ms
	99%	time for request:	73.28ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```

#### RPC WRITE THREE NODES BENCHMARK
```
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	100000
	Total time:	0.97s
	Requests per second:	103035.21
	Fastest time for request:	8.91ms
	Average time per request:	38.82ms
	Slowest time for request:	88.05ms

Time:
	10%	time for request:	21.52ms
	50%	time for request:	38.90ms
	90%	time for request:	54.88ms
	99%	time for request:	76.74ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```

#### ETCD WRITE THREE NODES BENCHMARK
```
Summary:
	Conns:	8
	Clients:	512
	Total calls:	100000
	Total time:	2.54s
	Requests per second:	39357.45
	Fastest time for request:	0.90ms
	Average time per request:	12.90ms
	Slowest time for request:	71.50ms

Time:
	10%	time for request:	7.10ms
	50%	time for request:	10.50ms
	90%	time for request:	17.50ms
	99%	time for request:	52.10ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```
### Licence
This package is licenced under a MIT licence (Copyright (c) 2019 Mort Huang)

### Authors
raft was written by Mort Huang.
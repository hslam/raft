# RAFT

## [Example raftdb](https://hslam.com/git/x/raft/src/master/example/raftdb "raftdb") Benchmark


### Mac Environment
* **CPU** 4 Cores 2.9 GHz
* **Memory** 8 GiB

#### Linux Environment
* **CPU** 12 Cores 3.1 GHz
* **Memory** 24 GiB

System  |Cluster    |Operation  |Transport  |TPS    |Fastest(ms) |Average(ms) |Slowest(ms)
 ---- | ----- | ------  | ------  | ------ | ------ | ------ | ------
Mac     |Singleton  |READINDEX  |HTTP   |10179  |8      |19     |77
Mac     |Singleton  |READINDEX  |RPC    |90342  |15     |43     |117
Mac     |Singleton  |WRITE      |HTTP   |10148  |8      |19     |80
Mac     |Singleton  |WRITE      |RPC    |52120  |14     |76     |217
Mac     |ThreeNodes |READINDEX  |HTTP   |4281   |13     |46     |90
Mac     |ThreeNodes |READINDEX  |RPC    |51943  |15     |75     |175
Mac     |ThreeNodes |WRITE      |HTTP   |4151   |10     |48     |104
Mac     |ThreeNodes |WRITE      |RPC    |35560  |28     |111    |306
Linux   |Singleton  |READINDEX  |HTTP   |74456  |2      |6      |110
Linux   |Singleton  |READINDEX  |RPC    |293865 |4      |13     |35
Linux   |Singleton  |WRITE      |HTTP   |57488  |2      |8      |119
Linux   |Singleton  |WRITE      |RPC    |132045 |6      |30     |86
Linux   |ThreeNodes |READINDEX  |HTTP   |43053  |2      |11     |1125
Linux   |ThreeNodes |READINDEX  |RPC    |267685 |4      |14     |44
Linux   |ThreeNodes |WRITE      |HTTP   |35241  |4      |14     |114
Linux   |ThreeNodes |WRITE      |RPC    |103035 |8      |38     |88


#### HTTP WRITE THREE NODES BENCHMARK
```
Summary:
	Clients:	200
	Parallel calls per client:	1
	Total calls:	100000
	Total time:	24.09s
	Requests per second:	4151.37
	Fastest time for request:	10.85ms
	Average time per request:	48.14ms
	Slowest time for request:	104.61ms

Time:
	10%	time for request:	38.17ms
	50%	time for request:	47.81ms
	90%	time for request:	57.93ms
	99%	time for request:	71.07ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```

#### RPC WRITE THREE NODES BENCHMARK
```
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	100000
	Total time:	2.81s
	Requests per second:	35560.85
	Fastest time for request:	28.83ms
	Average time per request:	111.83ms
	Slowest time for request:	306.85ms

Time:
	10%	time for request:	65.54ms
	50%	time for request:	107.06ms
	90%	time for request:	163.19ms
	99%	time for request:	255.75ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```
#### ETCD WRITE THREE NODES BENCHMARK
```
Summary:
	Conns:	8
	Clients:	512
	Total calls:	100000
	Total time:	8.69s
	Requests per second:	11497.77
	Fastest time for request:	9.10ms
	Average time per request:	44.20ms
	Slowest time for request:	133.40ms

Time:
	10%	time for request:	31.30ms
	50%	time for request:	42.30ms
	90%	time for request:	58.90ms
	99%	time for request:	80.20ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
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




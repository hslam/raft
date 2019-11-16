# RAFT

## [Example raftdb](https://hslam.com/git/x/raft/src/master/example/raftdb "raftdb") Benchmark
### Mac Environment
* **CPU** 4 Cores 2.9 GHz
* **Memory** 8 GiB

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
#### Linux Environment
* **CPU** 12 Cores 3.1 GHz
* **Memory** 24 GiB

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
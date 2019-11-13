# RAFT

## [Raftdb](https://hslam.com/git/x/raft/src/master/example/raftdb "raftdb") Benchmark
### Mac Environment
* **CPU** 4 Cores 2.9 GHz
* **Memory** 8 GiB

#### HTTP WRITE THREE NODES BENCHMARK
```
Summary:
	Clients:	200
	Parallel calls per client:	1
	Total calls:	100000
	Total time:	26.41s
	Requests per second:	3786.26
	Fastest time for request:	14.56ms
	Average time per request:	52.77ms
	Slowest time for request:	149.84ms

Time:
	1%	time for request:	31.95ms
	50%	time for request:	52.46ms
	99%	time for request:	77.31ms

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
	Total time:	2.79s
	Requests per second:	35870.73
	Fastest time for request:	35.97ms
	Average time per request:	109.79ms
	Slowest time for request:	286.30ms

Time:
	1%	time for request:	50.66ms
	50%	time for request:	103.42ms
	99%	time for request:	226.59ms

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
	Total calls:	1000000
	Total time:	27.96s
	Requests per second:	35760.43
	Fastest time for request:	4.19ms
	Average time per request:	14.29ms
	Slowest time for request:	89.11ms

Time:
	1%	time for request:	6.59ms
	50%	time for request:	10.29ms
	99%	time for request:	44.19ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```

#### RPC WRITE THREE NODES BENCHMARK
```
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	1000000
	Total time:	9.95s
	Requests per second:	100543.87
	Fastest time for request:	7.64ms
	Average time per request:	40.63ms
	Slowest time for request:	221.84ms

Time:
	1%	time for request:	15.78ms
	50%	time for request:	37.78ms
	99%	time for request:	138.01ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```
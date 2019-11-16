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

### Define
* **trans**     transport
* **tps**     requests per second
* **fast**     fastest(ms)
* **med**     median(ms)
* **avg**     average(ms)
* **p99**     99th percentile time (ms)
* **slow**     slowest(ms)


```
system  cluster     operation   trans   tps     avg     fast    med     p99     slow
Mac     Singleton   READINDEX   HTTP    10179   19.62   8.33    19.19   39.07   77.90
Mac     Singleton   READINDEX   RPC     90342   43.93   15.64   37.03   103.13  117.00
Mac     Singleton   WRITE       HTTP    10148   19.69   8.04    19.22   40.15   80.52
Mac     Singleton   WRITE       RPC     52120   76.24   14.32   68.16   178.10  217.68
Mac     ThreeNodes  READINDEX   HTTP    4281    13.39   46.96   46.67   67.34   90.54
Mac     ThreeNodes  READINDEX   RPC     51943   75.95   15.47   64.37   169.92  175.86
Mac     ThreeNodes  WRITE       HTTP    4151    48.14   10.85   47.81   71.07   104.61
Mac     ThreeNodes  WRITE       RPC     35560   111.83  28.83   107.06  255.75  306.85
Linux   Singleton   READINDEX   HTTP    74456   6.63    2.62    6.23    12.12   110.90
Linux   Singleton   READINDEX   RPC     293865  13.14   4.09    12.14   31.22   35.09
Linux   Singleton   WRITE       HTTP    57488   8.79    2.19    7.68    24.00   119.71
Linux   Singleton   WRITE       RPC     132045  30.21   6.39    27.59   70.14   86.11
Linux   ThreeNodes  READINDEX   HTTP    43053   11.72   2.83    7.43    58.92   1125.58
Linux   ThreeNodes  READINDEX   RPC     267685  14.65   4.36    13.47   31.59   44.72
Linux   ThreeNodes  WRITE       HTTP    35241   14.42   4.21    10.41   73.28   114.84
Linux   ThreeNodes  WRITE       RPC     103035  38.82   8.91    38.90   76.74   88.05
```
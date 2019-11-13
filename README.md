# RAFT

## example raftdb benchmark

### Mac Environment
* **CPU** 4 Cores 2.9 GHz
* **Memory** 8 GiB

#### HTTP READINDEX SINGLETON BENCHMARK
./http_read_index -p=7001 -parallel=1 -total=100000 -clients=200 -bar=false
```
Summary:
	Clients:	200
	Parallel calls per client:	1
	Total calls:	100000
	Total time:	9.85s
	Requests per second:	10150.97
	Fastest time for request:	8.18ms
	Average time per request:	19.68ms
	Slowest time for request:	57.69ms

Time:
	0.1%	time for request:	10.74ms
	1%	time for request:	12.44ms
	5%	time for request:	14.04ms
	10%	time for request:	15.09ms
	25%	time for request:	17.16ms
	50%	time for request:	19.52ms
	75%	time for request:	21.73ms
	90%	time for request:	24.05ms
	95%	time for request:	25.83ms
	99%	time for request:	30.83ms
	99.9%	time for request:	44.83ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```
#### RPC READINDEX SINGLETON BENCHMARK
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=true -batch=true -batch_async=true -clients=8
```
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	100000
	Total time:	1.10s
	Requests per second:	90934.30
	Fastest time for request:	15.10ms
	Average time per request:	43.59ms
	Slowest time for request:	93.56ms

Time:
	0.1%	time for request:	16.30ms
	1%	time for request:	19.96ms
	5%	time for request:	23.74ms
	10%	time for request:	25.25ms
	25%	time for request:	29.21ms
	50%	time for request:	39.08ms
	75%	time for request:	55.79ms
	90%	time for request:	66.65ms
	95%	time for request:	73.91ms
	99%	time for request:	85.80ms
	99.9%	time for request:	91.39ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```
#### HTTP WRITE SINGLETON BENCHMARK
./http -p=7001 -parallel=1 -total=100000 -clients=200 -bar=false
```
Summary:
	Clients:	200
	Parallel calls per client:	1
	Total calls:	100000
	Total time:	10.22s
	Requests per second:	9788.85
	Fastest time for request:	8.13ms
	Average time per request:	20.38ms
	Slowest time for request:	124.21ms

Time:
	0.1%	time for request:	10.06ms
	1%	time for request:	11.65ms
	5%	time for request:	13.55ms
	10%	time for request:	14.68ms
	25%	time for request:	16.80ms
	50%	time for request:	19.43ms
	75%	time for request:	22.18ms
	90%	time for request:	25.95ms
	95%	time for request:	29.45ms
	99%	time for request:	44.62ms
	99.9%	time for request:	106.39ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```
#### RPC WRITE SINGLETON BENCHMARK
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=true -batch=true -batch_async=true -clients=8
```
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	100000
	Total time:	1.89s
	Requests per second:	52974.04
	Fastest time for request:	18.12ms
	Average time per request:	75.38ms
	Slowest time for request:	273.51ms

Time:
	0.1%	time for request:	22.84ms
	1%	time for request:	31.76ms
	5%	time for request:	37.04ms
	10%	time for request:	41.67ms
	25%	time for request:	50.31ms
	50%	time for request:	66.73ms
	75%	time for request:	89.67ms
	90%	time for request:	112.47ms
	95%	time for request:	142.17ms
	99%	time for request:	230.26ms
	99.9%	time for request:	259.16ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```

#### HTTP READINDEX THREE NODES BENCHMARK
./http_read_index -p=7001 -parallel=1 -total=100000 -clients=200 -bar=false
```
Summary:
	Clients:	200
	Parallel calls per client:	1
	Total calls:	100000
	Total time:	25.28s
	Requests per second:	3956.42
	Fastest time for request:	12.10ms
	Average time per request:	50.47ms
	Slowest time for request:	154.46ms

Time:
	0.1%	time for request:	20.75ms
	1%	time for request:	28.66ms
	5%	time for request:	36.06ms
	10%	time for request:	38.65ms
	25%	time for request:	43.62ms
	50%	time for request:	50.25ms
	75%	time for request:	56.30ms
	90%	time for request:	62.14ms
	95%	time for request:	66.46ms
	99%	time for request:	75.99ms
	99.9%	time for request:	132.55ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```

#### RPC READINDEX THREE NODES BENCHMARK
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=true -batch=true -batch_async=true -clients=8
```
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	100000
	Total time:	2.61s
	Requests per second:	38273.69
	Fastest time for request:	31.67ms
	Average time per request:	104.97ms
	Slowest time for request:	467.87ms

Time:
	0.1%	time for request:	33.10ms
	1%	time for request:	34.60ms
	5%	time for request:	42.70ms
	10%	time for request:	49.92ms
	25%	time for request:	60.70ms
	50%	time for request:	84.53ms
	75%	time for request:	120.91ms
	90%	time for request:	149.47ms
	95%	time for request:	283.86ms
	99%	time for request:	429.59ms
	99.9%	time for request:	463.28ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```

#### HTTP WRITE THREE NODES BENCHMARK
./http -p=7001 -parallel=1 -total=100000 -clients=200 -bar=false
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
	0.1%	time for request:	26.49ms
	1%	time for request:	31.95ms
	5%	time for request:	38.27ms
	10%	time for request:	41.12ms
	25%	time for request:	46.21ms
	50%	time for request:	52.46ms
	75%	time for request:	58.59ms
	90%	time for request:	64.38ms
	95%	time for request:	67.82ms
	99%	time for request:	77.31ms
	99.9%	time for request:	138.84ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```

#### RPC WRITE THREE NODES BENCHMARK
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=true -batch=true -batch_async=true -clients=8
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
	0.1%	time for request:	40.61ms
	1%	time for request:	50.66ms
	5%	time for request:	59.55ms
	10%	time for request:	64.28ms
	25%	time for request:	76.24ms
	50%	time for request:	103.42ms
	75%	time for request:	134.57ms
	90%	time for request:	163.96ms
	95%	time for request:	195.88ms
	99%	time for request:	226.59ms
	99.9%	time for request:	244.37ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```

#### Linux Environment
* **CPU** 12 Cores 3.1 GHz
* **Memory** 24 GiB

#### HTTP READINDEX SINGLETON BENCHMARK
./http_read_index -p=7001 -parallel=1 -total=100000 -clients=512 -bar=false
```
Summary:
	Clients:	512
	Parallel calls per client:	1
	Total calls:	1000000
	Total time:	12.32s
	Requests per second:	81156.33
	Fastest time for request:	2.14ms
	Average time per request:	6.28ms
	Slowest time for request:	1092.09ms

Time:
	0.1%	time for request:	3.15ms
	1%	time for request:	3.82ms
	5%	time for request:	4.24ms
	10%	time for request:	4.49ms
	25%	time for request:	5.04ms
	50%	time for request:	5.95ms
	75%	time for request:	6.77ms
	90%	time for request:	7.60ms
	95%	time for request:	8.40ms
	99%	time for request:	12.19ms
	99.9%	time for request:	31.16ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```
#### RPC READINDEX SINGLETON BENCHMARK
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=1000000 -multiplexing=true -batch=true -batch_async=true -clients=8
```
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	1000000
	Total time:	3.16s
	Requests per second:	316606.89
	Fastest time for request:	3.70ms
	Average time per request:	12.87ms
	Slowest time for request:	83.28ms

Time:
	0.1%	time for request:	4.69ms
	1%	time for request:	5.62ms
	5%	time for request:	6.65ms
	10%	time for request:	7.16ms
	25%	time for request:	8.26ms
	50%	time for request:	11.05ms
	75%	time for request:	15.28ms
	90%	time for request:	18.34ms
	95%	time for request:	21.80ms
	99%	time for request:	54.32ms
	99.9%	time for request:	77.35ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```
#### HTTP WRITE SINGLETON BENCHMARK
./http -p=7001 -parallel=1 -total=100000 -clients=512 -bar=false
```
Summary:
	Clients:	512
	Parallel calls per client:	1
	Total calls:	1000000
	Total time:	14.59s
	Requests per second:	68521.77
	Fastest time for request:	2.21ms
	Average time per request:	7.46ms
	Slowest time for request:	94.04ms

Time:
	0.1%	time for request:	3.26ms
	1%	time for request:	3.97ms
	5%	time for request:	4.72ms
	10%	time for request:	5.10ms
	25%	time for request:	5.74ms
	50%	time for request:	6.61ms
	75%	time for request:	8.08ms
	90%	time for request:	10.15ms
	95%	time for request:	12.07ms
	99%	time for request:	23.46ms
	99.9%	time for request:	48.46ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```
#### RPC WRITE SINGLETON BENCHMARK
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=1000000 -multiplexing=true -batch=true -batch_async=true -clients=8
```
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	1000000
	Total time:	7.32s
	Requests per second:	136626.77
	Fastest time for request:	2.44ms
	Average time per request:	29.89ms
	Slowest time for request:	214.66ms

Time:
	0.1%	time for request:	7.30ms
	1%	time for request:	10.49ms
	5%	time for request:	13.93ms
	10%	time for request:	16.21ms
	25%	time for request:	20.87ms
	50%	time for request:	27.24ms
	75%	time for request:	34.34ms
	90%	time for request:	44.10ms
	95%	time for request:	53.76ms
	99%	time for request:	93.03ms
	99.9%	time for request:	161.42ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```
#### HTTP READINDEX THREE NODES BENCHMARK
./http_read_index -p=7001 -parallel=1 -total=100000 -clients=512 -bar=false
```
Summary:
	Clients:	512
	Parallel calls per client:	1
	Total calls:	1000000
	Total time:	21.79s
	Requests per second:	45889.22
	Fastest time for request:	2.51ms
	Average time per request:	11.14ms
	Slowest time for request:	1234.28ms

Time:
	0.1%	time for request:	3.69ms
	1%	time for request:	4.36ms
	5%	time for request:	5.18ms
	10%	time for request:	5.70ms
	25%	time for request:	6.73ms
	50%	time for request:	8.09ms
	75%	time for request:	9.19ms
	90%	time for request:	21.57ms
	95%	time for request:	40.95ms
	99%	time for request:	48.76ms
	99.9%	time for request:	52.49ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```

#### RPC READINDEX THREE NODES BENCHMARK
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=1000000 -multiplexing=true -batch=true -batch_async=true -clients=8
```
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	1000000
	Total time:	4.37s
	Requests per second:	228681.30
	Fastest time for request:	4.26ms
	Average time per request:	17.84ms
	Slowest time for request:	99.61ms

Time:
	0.1%	time for request:	5.20ms
	1%	time for request:	7.01ms
	5%	time for request:	8.05ms
	10%	time for request:	8.78ms
	25%	time for request:	10.05ms
	50%	time for request:	13.42ms
	75%	time for request:	20.08ms
	90%	time for request:	34.50ms
	95%	time for request:	45.92ms
	99%	time for request:	64.38ms
	99.9%	time for request:	80.49ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```

#### HTTP WRITE THREE NODES BENCHMARK
./http -p=7001 -parallel=1 -total=100000 -clients=512 -bar=false
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
	0.1%	time for request:	5.52ms
	1%	time for request:	6.59ms
	5%	time for request:	7.59ms
	10%	time for request:	8.12ms
	25%	time for request:	9.03ms
	50%	time for request:	10.29ms
	75%	time for request:	12.76ms
	90%	time for request:	35.15ms
	95%	time for request:	39.80ms
	99%	time for request:	44.19ms
	99.9%	time for request:	56.68ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```

#### RPC WRITE THREE NODES BENCHMARK
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=1000000 -multiplexing=true -batch=true -batch_async=true -clients=8
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
	0.1%	time for request:	11.98ms
	1%	time for request:	15.78ms
	5%	time for request:	19.42ms
	10%	time for request:	21.87ms
	25%	time for request:	28.15ms
	50%	time for request:	37.78ms
	75%	time for request:	46.18ms
	90%	time for request:	57.91ms
	95%	time for request:	72.67ms
	99%	time for request:	138.01ms
	99.9%	time for request:	196.33ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```
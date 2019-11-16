Example raftdb

build
```
go build -tags=use_cgo main.go

```
build for linux
```
docker run --rm -v $GOPATH/src:/go/src golang:1.12 bash -c 'cd $GOPATH/src/hslam.com/git/x/raft/example/raftdb && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags=use_cgo main.go'
```

Singleton
```sh
./raftdb -h=localhost -p=7001 -c=8001 -f=9001 -d=6061 -m=8 -peers="" -path=./raftdb.1
```
Three nodes
```sh
./raftdb -h=localhost -p=7001 -c=8001 -f=9001 -d=6061 -m=8 -peers=localhost:9001,localhost:9002,localhost:9003 -path=./raftdb.1
./raftdb -h=localhost -p=7002 -c=8002 -f=9002 -d=6062 -m=8 -peers=localhost:9001,localhost:9002,localhost:9003 -path=./raftdb.2
./raftdb -h=localhost -p=7003 -c=8003 -f=9003 -d=6063 -m=8 -peers=localhost:9001,localhost:9002,localhost:9003 -path=./raftdb.3
```
Http Set
```
curl -XPOST http://localhost:7001/db/foo -d 'bar'
```
Http Get
```
curl http://localhost:7001/db/foo
```

Benchmark
```
#!/bin/sh

nohup ./raftdb -h=localhost -p=7001 -c=8001 -f=9001 -d=6061 -m=8 -peers="" -path=./tmp/default.raftdb.1  >> ./tmp/default.out.1.log 2>&1 &
sleep 30s
curl -XPOST http://localhost:7001/db/foo -d 'bar'
sleep 3s
curl http://localhost:7001/db/foo
sleep 3s
./http_read_index -p=7001 -parallel=1 -total=100000 -clients=200 -bar=false
sleep 10s
./rpc_read_index -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=true -batch=true -batch_async=true -clients=8 -bar=false
sleep 10s
./http_write -p=7001 -parallel=1 -total=100000 -clients=200 -bar=false
sleep 10s
./rpc_write -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=true -batch=true -batch_async=true -clients=8 -bar=false
sleep 10s

killall raftdb

sleep 3s

nohup ./raftdb -h=localhost -p=7001 -c=8001 -f=9001 -d=6061 -m=8 -peers=localhost:9001,localhost:9002,localhost:9003 -path=./tmp/raftdb.1  >> ./tmp/out.1.log 2>&1 &
sleep 3s
nohup ./raftdb -h=localhost -p=7002 -c=8002 -f=9002 -d=6062 -m=8 -peers=localhost:9001,localhost:9002,localhost:9003 -path=./tmp/raftdb.2  >> ./tmp/out.2.log 2>&1 &
nohup ./raftdb -h=localhost -p=7003 -c=8003 -f=9003 -d=6063 -m=8 -peers=localhost:9001,localhost:9002,localhost:9003 -path=./tmp/raftdb.3  >> ./tmp/out.3.log 2>&1 &
sleep 30s
curl -XPOST http://localhost:7001/db/foo -d 'bar'
sleep 3s
curl http://localhost:7001/db/foo
sleep 3s
./http_read_index -p=7001 -parallel=1 -total=100000 -clients=200 -bar=false
sleep 10s
./rpc_read_index -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=true -batch=true -batch_async=true -clients=8 -bar=false
sleep 10s
./http_write -p=7001 -parallel=1 -total=100000 -clients=200 -bar=false
sleep 10s
./rpc_write -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=true -batch=true -batch_async=true -clients=8 -bar=false
sleep 10s

killall raftdb
sleep 3s
```

### Mac Environment
* **CPU** 4 Cores 2.9 GHz
* **Memory** 8 GiB

#### HTTP READINDEX SINGLETON BENCHMARK
```
Summary:
	Clients:	200
	Parallel calls per client:	1
	Total calls:	100000
	Total time:	9.82s
	Requests per second:	10179.94
	Fastest time for request:	8.33ms
	Average time per request:	19.62ms
	Slowest time for request:	77.90ms

Time:
	0.1%	time for request:	10.59ms
	1%	time for request:	12.01ms
	5%	time for request:	13.63ms
	10%	time for request:	14.66ms
	25%	time for request:	16.71ms
	50%	time for request:	19.19ms
	75%	time for request:	21.52ms
	90%	time for request:	24.35ms
	95%	time for request:	26.69ms
	99%	time for request:	39.07ms
	99.9%	time for request:	53.84ms

Result:
	Response ok:	99990 (99.99%)
	Errors:	10 (0.01%)
```
#### RPC READINDEX SINGLETON BENCHMARK
```
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	100000
	Total time:	1.11s
	Requests per second:	90342.32
	Fastest time for request:	15.64ms
	Average time per request:	43.93ms
	Slowest time for request:	117.00ms

Time:
	0.1%	time for request:	17.03ms
	1%	time for request:	21.91ms
	5%	time for request:	24.42ms
	10%	time for request:	25.86ms
	25%	time for request:	29.73ms
	50%	time for request:	37.03ms
	75%	time for request:	56.18ms
	90%	time for request:	65.65ms
	95%	time for request:	77.28ms
	99%	time for request:	103.13ms
	99.9%	time for request:	115.64ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```
#### HTTP WRITE SINGLETON BENCHMARK
```
Summary:
	Clients:	200
	Parallel calls per client:	1
	Total calls:	100000
	Total time:	9.85s
	Requests per second:	10148.56
	Fastest time for request:	8.04ms
	Average time per request:	19.69ms
	Slowest time for request:	80.52ms

Time:
	0.1%	time for request:	10.33ms
	1%	time for request:	11.61ms
	5%	time for request:	13.40ms
	10%	time for request:	14.57ms
	25%	time for request:	16.66ms
	50%	time for request:	19.22ms
	75%	time for request:	21.64ms
	90%	time for request:	24.17ms
	95%	time for request:	26.31ms
	99%	time for request:	40.15ms
	99.9%	time for request:	67.27ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```
#### RPC WRITE SINGLETON BENCHMARK
```
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	100000
	Total time:	1.92s
	Requests per second:	52120.31
	Fastest time for request:	14.32ms
	Average time per request:	76.24ms
	Slowest time for request:	217.68ms

Time:
	0.1%	time for request:	20.70ms
	1%	time for request:	29.58ms
	5%	time for request:	38.58ms
	10%	time for request:	43.33ms
	25%	time for request:	53.16ms
	50%	time for request:	68.16ms
	75%	time for request:	92.40ms
	90%	time for request:	119.66ms
	95%	time for request:	143.28ms
	99%	time for request:	178.10ms
	99.9%	time for request:	198.92ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```

#### HTTP READINDEX THREE NODES BENCHMARK
```
Summary:
	Clients:	200
	Parallel calls per client:	1
	Total calls:	100000
	Total time:	23.36s
	Requests per second:	4281.05
	Fastest time for request:	13.39ms
	Average time per request:	46.67ms
	Slowest time for request:	90.54ms

Time:
	0.1%	time for request:	17.19ms
	1%	time for request:	24.86ms
	5%	time for request:	33.15ms
	10%	time for request:	35.96ms
	25%	time for request:	40.63ms
	50%	time for request:	46.96ms
	75%	time for request:	52.49ms
	90%	time for request:	57.01ms
	95%	time for request:	60.56ms
	99%	time for request:	67.34ms
	99.9%	time for request:	77.95ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)

```

#### RPC READINDEX THREE NODES BENCHMARK
```
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	100000
	Total time:	1.93s
	Requests per second:	51943.17
	Fastest time for request:	15.47ms
	Average time per request:	75.95ms
	Slowest time for request:	175.86ms

Time:
	0.1%	time for request:	18.25ms
	1%	time for request:	21.86ms
	5%	time for request:	35.67ms
	10%	time for request:	41.07ms
	25%	time for request:	52.78ms
	50%	time for request:	64.37ms
	75%	time for request:	101.10ms
	90%	time for request:	116.75ms
	95%	time for request:	143.79ms
	99%	time for request:	169.92ms
	99.9%	time for request:	175.83ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)```
```
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
	0.1%	time for request:	23.20ms
	1%	time for request:	30.86ms
	5%	time for request:	35.67ms
	10%	time for request:	38.17ms
	25%	time for request:	42.44ms
	50%	time for request:	47.81ms
	75%	time for request:	53.37ms
	90%	time for request:	57.93ms
	95%	time for request:	61.26ms
	99%	time for request:	71.07ms
	99.9%	time for request:	95.43ms

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
	0.1%	time for request:	38.97ms
	1%	time for request:	50.30ms
	5%	time for request:	58.22ms
	10%	time for request:	65.54ms
	25%	time for request:	77.74ms
	50%	time for request:	107.06ms
	75%	time for request:	133.80ms
	90%	time for request:	163.19ms
	95%	time for request:	193.73ms
	99%	time for request:	255.75ms
	99.9%	time for request:	306.77ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```

#### Linux Environment
* **CPU** 12 Cores 3.1 GHz
* **Memory** 24 GiB

#### HTTP READINDEX SINGLETON BENCHMARK
```
Summary:
	Clients:	512
	Parallel calls per client:	1
	Total calls:	100000
	Total time:	1.34s
	Requests per second:	74456.78
	Fastest time for request:	2.62ms
	Average time per request:	6.63ms
	Slowest time for request:	110.90ms

Time:
	0.1%	time for request:	3.53ms
	1%	time for request:	4.05ms
	5%	time for request:	4.44ms
	10%	time for request:	4.69ms
	25%	time for request:	5.27ms
	50%	time for request:	6.23ms
	75%	time for request:	7.13ms
	90%	time for request:	8.12ms
	95%	time for request:	9.27ms
	99%	time for request:	12.12ms
	99.9%	time for request:	81.47ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```
#### RPC READINDEX SINGLETON BENCHMARK
```
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	100000
	Total time:	0.34s
	Requests per second:	293865.27
	Fastest time for request:	4.09ms
	Average time per request:	13.14ms
	Slowest time for request:	35.09ms

Time:
	0.1%	time for request:	5.12ms
	1%	time for request:	5.63ms
	5%	time for request:	6.82ms
	10%	time for request:	7.24ms
	25%	time for request:	8.30ms
	50%	time for request:	12.14ms
	75%	time for request:	16.22ms
	90%	time for request:	20.89ms
	95%	time for request:	26.36ms
	99%	time for request:	31.22ms
	99.9%	time for request:	34.33ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```
#### HTTP WRITE SINGLETON BENCHMARK
```
Summary:
	Clients:	512
	Parallel calls per client:	1
	Total calls:	100000
	Total time:	1.74s
	Requests per second:	57488.54
	Fastest time for request:	2.19ms
	Average time per request:	8.79ms
	Slowest time for request:	119.71ms

Time:
	0.1%	time for request:	3.31ms
	1%	time for request:	3.70ms
	5%	time for request:	4.39ms
	10%	time for request:	5.05ms
	25%	time for request:	6.16ms
	50%	time for request:	7.68ms
	75%	time for request:	9.84ms
	90%	time for request:	12.58ms
	95%	time for request:	15.76ms
	99%	time for request:	24.00ms
	99.9%	time for request:	85.99ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```
#### RPC WRITE SINGLETON BENCHMARK
```
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	100000
	Total time:	0.76s
	Requests per second:	132045.80
	Fastest time for request:	6.39ms
	Average time per request:	30.21ms
	Slowest time for request:	86.11ms

Time:
	0.1%	time for request:	8.53ms
	1%	time for request:	11.02ms
	5%	time for request:	14.28ms
	10%	time for request:	16.39ms
	25%	time for request:	20.35ms
	50%	time for request:	27.59ms
	75%	time for request:	36.17ms
	90%	time for request:	49.05ms
	95%	time for request:	57.28ms
	99%	time for request:	70.14ms
	99.9%	time for request:	76.14ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```
#### HTTP READINDEX THREE NODES BENCHMARK
```
Summary:
    Clients:	512
	Parallel calls per client:	1
	Total calls:	100000
	Total time:	2.32s
	Requests per second:	43053.90
	Fastest time for request:	2.83ms
	Average time per request:	11.72ms
	Slowest time for request:	1125.58ms

Time:
	0.1%	time for request:	3.54ms
	1%	time for request:	4.09ms
	5%	time for request:	4.76ms
	10%	time for request:	5.24ms
	25%	time for request:	6.28ms
	50%	time for request:	7.43ms
	75%	time for request:	8.65ms
	90%	time for request:	12.77ms
	95%	time for request:	39.88ms
	99%	time for request:	58.92ms
	99.9%	time for request:	1068.66ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```

#### RPC READINDEX THREE NODES BENCHMARK
```
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	100000
	Total time:	0.37s
	Requests per second:	267685.30
	Fastest time for request:	4.36ms
	Average time per request:	14.65ms
	Slowest time for request:	44.72ms

Time:
	0.1%	time for request:	5.34ms
	1%	time for request:	6.50ms
	5%	time for request:	7.75ms
	10%	time for request:	8.26ms
	25%	time for request:	9.69ms
	50%	time for request:	13.47ms
	75%	time for request:	18.37ms
	90%	time for request:	22.27ms
	95%	time for request:	25.10ms
	99%	time for request:	31.59ms
	99.9%	time for request:	44.32ms

Result:
	Response ok:	1000000 (100.00%)
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
	0.1%	time for request:	5.59ms
	1%	time for request:	6.48ms
	5%	time for request:	7.32ms
	10%	time for request:	7.84ms
	25%	time for request:	8.85ms
	50%	time for request:	10.41ms
	75%	time for request:	13.18ms
	90%	time for request:	24.42ms
	95%	time for request:	42.78ms
	99%	time for request:	73.28ms
	99.9%	time for request:	108.08ms

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
	0.1%	time for request:	9.53ms
	1%	time for request:	14.58ms
	5%	time for request:	18.81ms
	10%	time for request:	21.52ms
	25%	time for request:	28.00ms
	50%	time for request:	38.90ms
	75%	time for request:	46.57ms
	90%	time for request:	54.88ms
	95%	time for request:	62.80ms
	99%	time for request:	76.74ms
	99.9%	time for request:	85.04ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```
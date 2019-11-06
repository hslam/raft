Example raftdb

Singleton
```sh
$ raftdb -h=localhost -p=7001 -c=8001 -f=9001 -path=~/raftdb.1
```
Three nodes
```sh
$ raftdb -h=localhost -p=7001 -c=8001 -f=9001 -peers=localhost:9001,localhost:9002,localhost:9003 -path=./raftdb.1
$ raftdb -h=localhost -p=7002 -c=8002 -f=9002 -peers=localhost:9001,localhost:9002,localhost:9003 -path=./raftdb.2
$ raftdb -h=localhost -p=7003 -c=8003 -f=9003 -peers=localhost:9001,localhost:9002,localhost:9003 -path=./raftdb.3
```

```
Set
curl -XPOST http://localhost:7001/db/foo -d 'bar'
```

```
Get
curl http://localhost:7001/db/foo
```

```
go-wrk -m="POST" -c=128 -t=4 -n=1000 -k=false -b="12345678901234567890123456789012" http://127.0.0.1:7003/db/meng
```

go-wrk -m="POST" -c=1 -t=1 -n=1000 -b="12345678901234567890123456789012" http://127.0.0.1:7003/db/meng

go-wrk -m="POST" -c=32 -t=4 -n=1000 -b="12345678901234567890123456789012" http://127.0.0.1:7003/db/meng



go-wrk -m="POST" -c=128 -t=4 -n=1000 -b="12345678901234567890123456789012" http://127.0.0.1:7003/db/meng


go-wrk -m="POST" -c=128 -t=4 -n=10000 -b="12345678901234567890123456789012" http://127.0.0.1:7001/db/meng

go-wrk -m="POST" -c=128 -t=4 -n=1000 -k=false -b="bar" http://127.0.0.1:7001/db/foo

go-wrk -m="POST" -c=1 -t=1 -n=1000 -k=false -b="bar" http://127.0.0.1:7001/db/foo


go-wrk -m="POST" -c=64 -t=1 -n=1000 -k=false -b="bar" http://127.0.0.1:7001/db/foo


curl -XPOST http://localhost:7003/db/foo -d 'bar'

curl http://localhost:7003/db/foo

go-wrk -m="POST" -c=1 -t=1 -n=1000 -k=false -b="bar" http://127.0.0.1:7001/db/foo

go-wrk -m="GET" -c=1 -t=1 -n=1000 -k=false http://localhost:7001/db/foo

go-wrk -m="POST" -c=128 -t=8 -n=10000 -k=false -b="bar" http://127.0.0.1:7001/db/foo

go-wrk -m="GET" -c=128 -t=8 -n=10000 -k=false http://localhost:7001/db/foo

go-wrk -m="POST" -c=1 -t=1 -n=1000 -k=false -b="bar" http://127.0.0.1:7003/db/foo


go-wrk -m="POST" -c=32 -t=8 -n=1000 -k=false -b="bar" http://127.0.0.1:7003/db/foo


go-wrk -m="POST" -c=128 -t=8 -n=1000 -k=false -b="bar" http://127.0.0.1:7003/db/foo


### HTTP BENCHMARK SINGLETON
```
Summary:
	Clients:	200
	Parallels:	1
	Total Calls:	100000
	Total time:	10.67s
	Requests per second:	9375.32
	Fastest time for request:	8.05ms
	Average time per request:	21.30ms
	Slowest time for request:	92.80ms

Time:
	0.1%	time for request:	10.29ms
	1%	time for request:	11.86ms
	5%	time for request:	13.74ms
	10%	time for request:	14.91ms
	25%	time for request:	17.19ms
	50%	time for request:	20.09ms
	75%	time for request:	23.45ms
	90%	time for request:	28.07ms
	95%	time for request:	33.33ms
	99%	time for request:	47.89ms
	99.9%	time for request:	80.68ms

Request:
	Total request body sizes:	3600000
	Average body size per request:	36.00 Byte
	Request rate per second:	337511.63 Byte/s (0.34 MByte/s)

Result:
	ResponseOk:	100000 (100.00%)
	Errors	:0 (0.00%)
```

### RPC BENCHMARK SINGLETON
```
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=false -batch=true -batch_async=true -clients=8
Summary:
	Clients:	8
	Parallels:	512
	Total Calls:	100000
	Total time:	1.90s
	Requests per second:	52579.39
	Fastest time for request:	13.46ms
	Average time per request:	75.74ms
	Slowest time for request:	219.71ms

Time:
	0.1%	time for request:	23.62ms
	1%	time for request:	29.38ms
	5%	time for request:	38.20ms
	10%	time for request:	42.36ms
	25%	time for request:	53.04ms
	50%	time for request:	68.43ms
	75%	time for request:	91.96ms
	90%	time for request:	122.98ms
	95%	time for request:	137.86ms
	99%	time for request:	156.67ms
	99.9%	time for request:	178.86ms

Request:
	Total request body sizes:	3600000
	Average body size per request:	36.00 Byte
	Request rate per second:	1892857.93 Byte/s (1.89 MByte/s)

Result:
	ResponseOk:	100000 (100.00%)
	Errors	:0 (0.00%)
```

### HTTP BENCHMARK THREE NODES
```
Summary:
	Clients:	200
	Parallels:	1
	Total Calls:	100000
	Total time:	25.35s
	Requests per second:	3944.01
	Fastest time for request:	15.02ms
	Average time per request:	50.65ms
	Slowest time for request:	115.46ms

Time:
	0.1%	time for request:	25.90ms
	1%	time for request:	31.57ms
	5%	time for request:	36.73ms
	10%	time for request:	39.21ms
	25%	time for request:	44.18ms
	50%	time for request:	50.35ms
	75%	time for request:	56.37ms
	90%	time for request:	62.15ms
	95%	time for request:	65.91ms
	99%	time for request:	74.59ms
	99.9%	time for request:	102.70ms

Request:
	Total request body sizes:	3600000
	Average body size per request:	36.00 Byte
	Request rate per second:	141984.30 Byte/s (0.14 MByte/s)

Result:
	ResponseOk:	100000 (100.00%)
	Errors	:0 (0.00%)
```

### RPC BENCHMARK THREE NODES
```
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=false -batch=true -batch_async=true -clients=8
Summary:
	Clients:	8
	Parallels:	512
	Total Calls:	100000
	Total time:	3.42s
	Requests per second:	29240.91
	Fastest time for request:	35.74ms
	Average time per request:	136.10ms
	Slowest time for request:	385.50ms

Time:
	0.1%	time for request:	42.59ms
	1%	time for request:	54.21ms
	5%	time for request:	64.44ms
	10%	time for request:	72.65ms
	25%	time for request:	88.43ms
	50%	time for request:	125.05ms
	75%	time for request:	163.59ms
	90%	time for request:	220.49ms
	95%	time for request:	263.67ms
	99%	time for request:	315.92ms
	99.9%	time for request:	341.29ms

Request:
	Total request body sizes:	3600000
	Average body size per request:	36.00 Byte
	Request rate per second:	1052672.82 Byte/s (1.05 MByte/s)

Result:
	ResponseOk:	100000 (100.00%)
	Errors	:0 (0.00%)
```
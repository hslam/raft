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
	Parallel calls per client:	1
	Total calls:	100000
	Total time:	10.17s
	Requests per second:	9832.32
	Fastest time for request:	7.88ms
	Average time per request:	20.31ms
	Slowest time for request:	91.06ms

Time:
	0.1%	time for request:	10.14ms
	1%	time for request:	11.63ms
	5%	time for request:	13.38ms
	10%	time for request:	14.47ms
	25%	time for request:	16.61ms
	50%	time for request:	19.27ms
	75%	time for request:	22.00ms
	90%	time for request:	25.84ms
	95%	time for request:	31.17ms
	99%	time for request:	48.74ms
	99.9%	time for request:	74.80ms

Request:
	Total request body sizes:	3600000
	Average body size per request:	36.00 Byte
	Request rate per second:	353963.65 Byte/s (0.35 MByte/s)

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```

### RPC BENCHMARK SINGLETON
```
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=true -batch=true -batch_async=true -clients=8
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	100000
	Total time:	1.98s
	Requests per second:	50527.33
	Fastest time for request:	18.84ms
	Average time per request:	77.87ms
	Slowest time for request:	259.58ms

Time:
	0.1%	time for request:	25.15ms
	1%	time for request:	31.12ms
	5%	time for request:	40.20ms
	10%	time for request:	44.53ms
	25%	time for request:	56.94ms
	50%	time for request:	74.17ms
	75%	time for request:	93.20ms
	90%	time for request:	116.09ms
	95%	time for request:	131.26ms
	99%	time for request:	165.62ms
	99.9%	time for request:	229.75ms

Request:
	Total request body sizes:	3600000
	Average body size per request:	36.00 Byte
	Request rate per second:	1818983.82 Byte/s (1.82 MByte/s)

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```

### HTTP BENCHMARK THREE NODES
```
Summary:
	Clients:	200
	Parallel calls per client:	1
	Total calls:	100000
	Total time:	27.36s
	Requests per second:	3655.64
	Fastest time for request:	15.02ms
	Average time per request:	54.66ms
	Slowest time for request:	109.34ms

Time:
	0.1%	time for request:	24.46ms
	1%	time for request:	29.96ms
	5%	time for request:	36.71ms
	10%	time for request:	39.95ms
	25%	time for request:	45.82ms
	50%	time for request:	53.44ms
	75%	time for request:	61.94ms
	90%	time for request:	71.21ms
	95%	time for request:	77.29ms
	99%	time for request:	88.61ms
	99.9%	time for request:	101.23ms

Request:
	Total request body sizes:	3600000
	Average body size per request:	36.00 Byte
	Request rate per second:	131602.98 Byte/s (0.13 MByte/s)

Result:
	Response ok:	99981 (99.98%)
	Errors:	19 (0.02%)
```

### RPC BENCHMARK THREE NODES
```
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=true -batch=true -batch_async=true -clients=8
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	100000
	Total time:	2.83s
	Requests per second:	35282.30
	Fastest time for request:	30.46ms
	Average time per request:	110.57ms
	Slowest time for request:	286.87ms

Time:
	0.1%	time for request:	38.73ms
	1%	time for request:	44.63ms
	5%	time for request:	53.53ms
	10%	time for request:	59.99ms
	25%	time for request:	72.50ms
	50%	time for request:	105.79ms
	75%	time for request:	136.79ms
	90%	time for request:	169.09ms
	95%	time for request:	206.31ms
	99%	time for request:	252.84ms
	99.9%	time for request:	277.27ms

Request:
	Total request body sizes:	3600000
	Average body size per request:	36.00 Byte
	Request rate per second:	1270162.95 Byte/s (1.27 MByte/s)

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```
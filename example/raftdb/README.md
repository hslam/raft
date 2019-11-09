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

Request:
	Total request body sizes:	3600000
	Average body size per request:	36.00 Byte
	Request rate per second:	352398.76 Byte/s (0.35 MByte/s)

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```

### RPC BENCHMARK SINGLETON
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

Request:
	Total request body sizes:	3600000
	Average body size per request:	36.00 Byte
	Request rate per second:	1907065.52 Byte/s (1.91 MByte/s)

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

Request:
	Total request body sizes:	3600000
	Average body size per request:	36.00 Byte
	Request rate per second:	136305.32 Byte/s (0.14 MByte/s)

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```

### RPC BENCHMARK THREE NODES
```
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=true -batch=true -batch_async=true -clients=8
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

Request:
	Total request body sizes:	3600000
	Average body size per request:	36.00 Byte
	Request rate per second:	1291346.40 Byte/s (1.29 MByte/s)

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```
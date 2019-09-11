

```sh
$ raftd -p 7001 ~/node.1
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

go-wrk -m="POST" -c=1 -t=1 -n=1000 -k=false -b="bar" http://127.0.0.1:7003/db/foo


go-wrk -m="POST" -c=32 -t=8 -n=1000 -k=false -b="bar" http://127.0.0.1:7003/db/foo


go-wrk -m="POST" -c=128 -t=8 -n=1000 -k=false -b="bar" http://127.0.0.1:7003/db/foo


### HTTP BENCHMARK SINGLETON
```
==========================BENCHMARK==========================
Used Connections:		1
Used Parallel Calls:		200
Total Number Of Calls:		100000

===========================TIMINGS===========================
Total time passed:		8.87s
Avg time per request:		17.71ms
Requests per second:		11268.45
Median time per request:	16.43ms
99th percentile time:		39.88ms
Slowest time for request:	114.00ms

=========================REQUESTDATA=========================
Total request body sizes:		3600000
Avg body size per request:		36.00 Byte
Transfer request rate per second:		405664.06 Byte/s (0.41 MByte/s)
==========================RESPONSES==========================
ResponseOk:			100000 (100.00%)
Errors:				0 (0.00%)
```

### RPC BENCHMARK SINGLETON

```
==========================BENCHMARK==========================
Used Connections:		16
Used Parallel Calls:		256
Total Number Of Calls:		1000000

===========================TIMINGS===========================
Total time passed:		32.69s
Avg time per request:		133.02ms
Requests per second:		30588.08
Median time per request:	111.34ms
99th percentile time:		370.92ms
Slowest time for request:	581.00ms

=========================REQUESTDATA=========================
Total request body sizes:		36000000
Avg body size per request:		36.00 Byte
Transfer request rate per second:		1101170.97 Byte/s (1.10 MByte/s)
==========================RESPONSES==========================
ResponseOk:			1000000 (100.00%)
Errors:				0 (0.00%)
```

### HTTP BENCHMARK THREE NODES
```
==========================BENCHMARK==========================
Used Connections:		1
Used Parallel Calls:		200
Total Number Of Calls:		100000

===========================TIMINGS===========================
Total time passed:		25.46s
Avg time per request:		50.78ms
Requests per second:		3927.23
Median time per request:	49.88ms
99th percentile time:		81.17ms
Slowest time for request:	199.00ms

=========================REQUESTDATA=========================
Total request body sizes:		3600000
Avg body size per request:		36.00 Byte
Transfer request rate per second:		141380.45 Byte/s (0.14 MByte/s)
==========================RESPONSES==========================
ResponseOk:			100000 (100.00%)
Errors:				0 (0.00%)
```

### RPC BENCHMARK THREE NODES
```
==========================BENCHMARK==========================
Used Connections:		16
Used Parallel Calls:		256
Total Number Of Calls:		1000001

===========================TIMINGS===========================
Total time passed:		57.58s
Avg time per request:		234.30ms
Requests per second:		17366.78
Median time per request:	194.53ms
99th percentile time:		818.00ms
Slowest time for request:	1271.00ms

=========================REQUESTDATA=========================
Total request body sizes:		36000036
Avg body size per request:		36.00 Byte
Transfer request rate per second:		625204.03 Byte/s (0.63 MByte/s)
==========================RESPONSES==========================
ResponseOk:			1000001 (100.00%)
Errors:				0 (0.00%)
```
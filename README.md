# RAFT

## example raftdb benchmark

### ENV

```
Mac 4 CPU 8 GiB
```

### SINGLETON HTTP
```
==========================BENCHMARK==========================
Used Connections:			200
Concurrent Calls Per Connection:	1
Total Number Of Calls:			100000

===========================TIMINGS===========================
Total time passed:			9.49s
Avg time per request:			18.93ms
Requests per second:			10539.60
Median time per request:		16.97ms
99th percentile time:			52.21ms
Slowest time for request:		159.00ms

=========================REQUESTDATA=========================
Total request body sizes:		3600000
Avg body size per request:		36.00 Byte
Transfer request rate per second:	379425.76 Byte/s (0.38 MByte/s)

==========================RESPONSES==========================
ResponseOk:				100000 (100.00%)
Errors:					0 (0.00%)
```

### SINGLETON RPC
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -total=1000000 -pipelining=true -multiplexing=false -batch=true -batch_async=true -clients=2
```
==========================BENCHMARK==========================
Used Connections:			2
Concurrent Calls Per Connection:	1024
Total Number Of Calls:			1000000

===========================TIMINGS===========================
Total time passed:			29.18s
Avg time per request:			59.56ms
Requests per second:			34265.14
Median time per request:		45.95ms
99th percentile time:			189.07ms
Slowest time for request:		335.00ms

=========================REQUESTDATA=========================
Total request body sizes:		36000000
Avg body size per request:		36.00 Byte
Transfer request rate per second:	1233545.19 Byte/s (1.23 MByte/s)

==========================RESPONSES==========================
ResponseOk:				1000000 (100.00%)
Errors:					0 (0.00%)
```

### THREE NODES HTTP
```
==========================BENCHMARK==========================
Used Connections:			200
Concurrent Calls Per Connection:	1
Total Number Of Calls:			100000

===========================TIMINGS===========================
Total time passed:			26.99s
Avg time per request:			53.86ms
Requests per second:			3704.42
Median time per request:		53.44ms
99th percentile time:			83.74ms
Slowest time for request:		118.00ms

=========================REQUESTDATA=========================
Total request body sizes:		3600000
Avg body size per request:		36.00 Byte
Transfer request rate per second:	133358.99 Byte/s (0.13 MByte/s)

==========================RESPONSES==========================
ResponseOk:				100000 (100.00%)
Errors:					0 (0.00%)
```

### THREE NODES RPC
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8002 -total=1000000 -pipelining=true -multiplexing=false -batch=true -batch_async=true -clients=2
```
==========================BENCHMARK==========================
Used Connections:			2
Concurrent Calls Per Connection:	1024
Total Number Of Calls:			1000000

===========================TIMINGS===========================
Total time passed:			66.42s
Avg time per request:			135.68ms
Requests per second:			15056.22
Median time per request:		123.75ms
99th percentile time:			328.28ms
Slowest time for request:		538.00ms

=========================REQUESTDATA=========================
Total request body sizes:		36000000
Avg body size per request:		36.00 Byte
Transfer request rate per second:	542023.94 Byte/s (0.54 MByte/s)

==========================RESPONSES==========================
ResponseOk:				1000000 (100.00%)
Errors:					0 (0.00%)
```
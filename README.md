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
Total time passed:			10.77s
Requests per second:			9285.83
Avg time per request:			21.48ms
Fastest time for request:		8.00ms
10% time for request:			14.89ms
25% time for request:			17.20ms
Median time per request:		20.08ms
75% time for request:			23.65ms
90% time for request:			29.21ms
99% time for request:			46.49ms
99.9% time for request:			96.86ms
Slowest time for request:		106.39ms

=========================REQUESTDATA=========================
Total request body sizes:		3600000
Avg body size per request:		36.00 Byte
Transfer request rate per second:	334289.84 Byte/s (0.33 MByte/s)

==========================RESPONSES==========================
ResponseOk:				100000 (100.00%)
Errors:					0 (0.00%)
```

### SINGLETON RPC
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=false -batch=true -batch_async=true -clients=8
```
==========================BENCHMARK==========================
Used Connections:			8
Concurrent Calls Per Connection:	512
Total Number Of Calls:			100000

===========================TIMINGS===========================
Total time passed:			2.17s
Requests per second:			46187.88
Avg time per request:			79.87ms
Fastest time for request:		23.16ms
10% time for request:			45.41ms
25% time for request:			56.49ms
Median time per request:		74.22ms
75% time for request:			96.28ms
90% time for request:			121.52ms
99% time for request:			185.21ms
99.9% time for request:			219.31ms
Slowest time for request:		237.31ms

=========================REQUESTDATA=========================
Total request body sizes:		3600000
Avg body size per request:		36.00 Byte
Transfer request rate per second:	1662763.79 Byte/s (1.66 MByte/s)

==========================RESPONSES==========================
ResponseOk:				100000 (100.00%)
Errors:					0 (0.00%)
```

### THREE NODES HTTP
```
==========================BENCHMARK==========================
Used Connections:			200
Concurrent Calls Per Connection:	1
Total Number Of Calls:			100000

===========================TIMINGS===========================
Total time passed:			26.37s
Requests per second:			3792.23
Avg time per request:			52.59ms
Fastest time for request:		13.96ms
10% time for request:			39.21ms
25% time for request:			44.67ms
Median time per request:		51.39ms
75% time for request:			58.79ms
90% time for request:			66.31ms
99% time for request:			87.11ms
99.9% time for request:			181.77ms
Slowest time for request:		198.13ms

=========================REQUESTDATA=========================
Total request body sizes:		3600000
Avg body size per request:		36.00 Byte
Transfer request rate per second:	136520.14 Byte/s (0.14 MByte/s)

==========================RESPONSES==========================
ResponseOk:				100000 (100.00%)
Errors:					0 (0.00%)
```

### THREE NODES RPC
./client -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=false -batch=true -batch_async=true -clients=8
```
==========================BENCHMARK==========================
Used Connections:			8
Concurrent Calls Per Connection:	512
Total Number Of Calls:			100000

===========================TIMINGS===========================
Total time passed:			3.09s
Requests per second:			32347.43
Avg time per request:			118.06ms
Fastest time for request:		37.31ms
10% time for request:			67.95ms
25% time for request:			82.10ms
Median time per request:		108.61ms
75% time for request:			146.92ms
90% time for request:			181.24ms
99% time for request:			244.84ms
99.9% time for request:			267.20ms
Slowest time for request:		333.80ms

=========================REQUESTDATA=========================
Total request body sizes:		3600000
Avg body size per request:		36.00 Byte
Transfer request rate per second:	1164507.37 Byte/s (1.16 MByte/s)

==========================RESPONSES==========================
ResponseOk:				100000 (100.00%)
Errors:					0 (0.00%)
```
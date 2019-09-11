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

```
==========================BENCHMARK==========================
Used Connections:			16
Concurrent Calls Per Connection:	256
Total Number Of Calls:			1000000

===========================TIMINGS===========================
Total time passed:			25.70s
Avg time per request:			104.68ms
Requests per second:			38914.09
Median time per request:		80.62ms
99th percentile time:			360.49ms
Slowest time for request:		621.00ms

=========================REQUESTDATA=========================
Total request body sizes:		36000000
Avg body size per request:		36.00 Byte
Transfer request rate per second:	1400907.18 Byte/s (1.40 MByte/s)

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
```
==========================BENCHMARK==========================
Used Connections:			16
Concurrent Calls Per Connection:	256
Total Number Of Calls:			1000000

===========================TIMINGS===========================
Total time passed:			52.86s
Avg time per request:			215.31ms
Requests per second:			18918.59
Median time per request:		181.93ms
99th percentile time:			630.44ms
Slowest time for request:		966.00ms

=========================REQUESTDATA=========================
Total request body sizes:		36000000
Avg body size per request:		36.00 Byte
Transfer request rate per second:	681069.30 Byte/s (0.68 MByte/s)

==========================RESPONSES==========================
ResponseOk:				1000000 (100.00%)
Errors:					0 (0.00%)
```
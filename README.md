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
Total time passed:			9.62s
Avg time per request:			19.18ms
Requests per second:			10396.80
Median time per request:		18.06ms
99th percentile time:			39.60ms
Slowest time for request:		158.00ms

=========================REQUESTDATA=========================
Total request body sizes:		3600000
Avg body size per request:		36.00 Byte
Transfer request rate per second:	374284.80 Byte/s (0.37 MByte/s)
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
Total time passed:			32.50s
Avg time per request:			132.05ms
Requests per second:			30771.57
Median time per request:		108.89ms
99th percentile time:			381.30ms
Slowest time for request:		574.00ms

=========================REQUESTDATA=========================
Total request body sizes:		36000000
Avg body size per request:		36.00 Byte
Transfer request rate per second:	1107776.40 Byte/s (1.11 MByte/s)
==========================RESPONSES==========================
ResponseOk:				1000000 (100.00%)
Errors:					0 (0.00%)
```

### THREE NODES HTTP
```
Used Connections:			200
Concurrent Calls Per Connection:	1
Total Number Of Calls:			100000

===========================TIMINGS===========================
Total time passed:			27.14s
Avg time per request:			54.10ms
Requests per second:			3684.92
Median time per request:		53.18ms
99th percentile time:			82.88ms
Slowest time for request:		120.00ms

=========================REQUESTDATA=========================
Total request body sizes:		3600000
Avg body size per request:		36.00 Byte
Transfer request rate per second:	132656.99 Byte/s (0.13 MByte/s)
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
Total time passed:			57.04s
Avg time per request:			232.33ms
Requests per second:			17531.26
Median time per request:		197.73ms
99th percentile time:			692.00ms
Slowest time for request:		949.00ms

=========================REQUESTDATA=========================
Total request body sizes:		36000000
Avg body size per request:		36.00 Byte
Transfer request rate per second:	631125.40 Byte/s (0.63 MByte/s)
==========================RESPONSES==========================
ResponseOk:				1000000 (100.00%)
Errors:					0 (0.00%)
```
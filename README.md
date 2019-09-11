
### env Mac 4 CPU 8 GiB

### HTTP BENCHMARK SINGLETON
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
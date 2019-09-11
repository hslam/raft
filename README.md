
### HTTP BENCHMARK
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

### RPC BENCHMARK

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
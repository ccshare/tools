### logrus
- perf
```
hey -c 300 -n 20000 http://127.0.0.1/logrus

Summary:
  Total:	1.8789 secs
  Slowest:	0.0690 secs
  Fastest:	0.0002 secs
  Average:	0.0280 secs
  Requests/sec:	10538.0678
  Total data:	633600 bytes
  Size/request:	32 bytes

Response time histogram:
  0.000 [1]	|
  0.007 [421]	|■
  0.014 [242]	|■
  0.021 [373]	|■
  0.028 [6048]	|■■■■■■■■■■■■■■■■■■■■
  0.035 [12355]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.041 [163]	|■
  0.048 [28]	|
  0.055 [38]	|
  0.062 [100]	|
  0.069 [31]	|

Latency distribution:
  10% in 0.0243 secs
  25% in 0.0270 secs
  50% in 0.0285 secs
  75% in 0.0305 secs
  90% in 0.0318 secs
  95% in 0.0325 secs
  99% in 0.0414 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0002 secs, 0.0002 secs, 0.0690 secs
  DNS-lookup:	0.0000 secs, 0.0000 secs, 0.0000 secs
  req write:	0.0000 secs, 0.0000 secs, 0.0052 secs
  resp wait:	0.0276 secs, 0.0001 secs, 0.0428 secs
  resp read:	0.0000 secs, 0.0000 secs, 0.0025 secs

Status code distribution:
  [200]	19800 responses
```
- log
```
time="2019-09-01T21:11:31+08:00" level=info msg="log1 sample info" index=0 name=walrus
time="2019-09-01T21:11:31+08:00" level=info msg="log1 sample info" index=1 name=walrus
time="2019-09-01T21:11:31+08:00" level=info msg="log1 sample info" index=2 name=walrus
time="2019-09-01T21:11:31+08:00" level=info msg="log1 sample info" index=3 name=walrus
```

### zap
- perf
```
hey -c 300 -n 20000 http://127.0.0.1/logzap

Summary:
  Total:	0.6334 secs
  Slowest:	0.0772 secs
  Fastest:	0.0001 secs
  Average:	0.0090 secs
  Requests/sec:	31262.0531
  Total data:	633600 bytes
  Size/request:	32 bytes

Response time histogram:
  0.000 [1]	    |
  0.008 [10233]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.016 [6797]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.023 [1929]	|■■■■■■■■
  0.031 [461]	|■■
  0.039 [263]	|■
  0.046 [52]	|
  0.054 [28]	|
  0.062 [23]	|
  0.070 [10]	|
  0.077 [3]	    |

Latency distribution:
  10% in 0.0012 secs
  25% in 0.0038 secs
  50% in 0.0076 secs
  75% in 0.0122 secs
  90% in 0.0173 secs
  95% in 0.0221 secs
  99% in 0.0356 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0001 secs, 0.0001 secs, 0.0772 secs
  DNS-lookup:	0.0000 secs, 0.0000 secs, 0.0000 secs
  req write:	0.0000 secs, 0.0000 secs, 0.0114 secs
  resp wait:	0.0033 secs, 0.0001 secs, 0.0755 secs
  resp read:	0.0032 secs, 0.0000 secs, 0.0281 secs

Status code distribution:
  [200]	19800 responses
```
- log
```
{"level":"info","ts":1567343504.275092,"caller":"logger/main.go:61","msg":"log2 sample info","index":0,"name":"walrus"}
{"level":"info","ts":1567343504.2751281,"caller":"logger/main.go:66","msg":"log2 sample info","index":1,"name":"walrus"}
{"level":"info","ts":1567343504.275136,"caller":"logger/main.go:71","msg":"log2 sample info","index":2,"name":"walrus"}
{"level":"info","ts":1567343504.2751412,"caller":"logger/main.go:76","msg":"log2 sample info","index":3,"name":"walrus"}
```
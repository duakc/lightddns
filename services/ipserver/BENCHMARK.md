```yaml
log:
  disabled: true

services:
  - enabled: true
    type: ipserver
    port: 8080
    dump: false
```

```shell
$ ab -m GET -k -n 10000 -c 100 "http://127.0.0.1:8080/?format=json"
```

```text
This is ApacheBench, Version 2.3 <$Revision: 1923142 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking 127.0.0.1 (be patient)


Server Software:        
Server Hostname:        127.0.0.1
Server Port:            8080

Document Path:          /?format=json
Document Length:        34 bytes

Concurrency Level:      100
Time taken for tests:   0.082 seconds
Complete requests:      10000
Failed requests:        0
Keep-Alive requests:    10000
Total transferred:      1750000 bytes
HTML transferred:       340000 bytes
Requests per second:    121749.29 [#/sec] (mean)
Time per request:       0.821 [ms] (mean)
Time per request:       0.008 [ms] (mean, across all concurrent requests)
Transfer rate:          20806.76 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.2      0       3
Processing:     0    1   0.3      1       7
Waiting:        0    1   0.3      1       4
Total:          0    1   0.4      1       7

Percentage of the requests served within a certain time (ms)
  50%      1
  66%      1
  75%      1
  80%      1
  90%      1
  95%      1
  98%      2
  99%      2
 100%      7 (longest request)

```

```shell
$ ab -m GET -k -n 10000 -c 100 "http://127.0.0.1:8080/?format=yaml"
```

```text
This is ApacheBench, Version 2.3 <$Revision: 1923142 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking 127.0.0.1 (be patient)


Server Software:        
Server Hostname:        127.0.0.1
Server Port:            8080

Document Path:          /?format=yaml
Document Length:        31 bytes

Concurrency Level:      100
Time taken for tests:   0.078 seconds
Complete requests:      10000
Failed requests:        0
Keep-Alive requests:    10000
Total transferred:      1720000 bytes
HTML transferred:       310000 bytes
Requests per second:    127599.85 [#/sec] (mean)
Time per request:       0.784 [ms] (mean)
Time per request:       0.008 [ms] (mean, across all concurrent requests)
Transfer rate:          21432.79 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.2      0       3
Processing:     0    1   0.4      1       4
Waiting:        0    1   0.4      1       3
Total:          0    1   0.5      1       6

Percentage of the requests served within a certain time (ms)
  50%      1
  66%      1
  75%      1
  80%      1
  90%      1
  95%      1
  98%      2
  99%      3
 100%      6 (longest request)
```

```shell
$ ab -m GET -k -n 10000 -c 100 "http://127.0.0.1:8080/"
```

```text
This is ApacheBench, Version 2.3 <$Revision: 1923142 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking 127.0.0.1 (be patient)


Server Software:        
Server Hostname:        127.0.0.1
Server Port:            8080

Document Path:          /
Document Length:        9 bytes

Concurrency Level:      100
Time taken for tests:   0.075 seconds
Complete requests:      10000
Failed requests:        0
Keep-Alive requests:    10000
Total transferred:      1490000 bytes
HTML transferred:       90000 bytes
Requests per second:    133315.56 [#/sec] (mean)
Time per request:       0.750 [ms] (mean)
Time per request:       0.008 [ms] (mean, across all concurrent requests)
Transfer rate:          19398.46 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.2      0       3
Processing:     0    1   0.3      1       4
Waiting:        0    1   0.3      1       3
Total:          0    1   0.5      1       5

Percentage of the requests served within a certain time (ms)
  50%      1
  66%      1
  75%      1
  80%      1
  90%      1
  95%      1
  98%      2
  99%      4
 100%      5 (longest request)
```

```shell
$ ab -m GET -k -n 10000 -c 100 -H "True-Client-IP: 1.1.1.1" "http://127.0.0.1:8080/"
```

```text
This is ApacheBench, Version 2.3 <$Revision: 1923142 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking 127.0.0.1 (be patient)


Server Software:        
Server Hostname:        127.0.0.1
Server Port:            8080

Document Path:          /
Document Length:        7 bytes

Concurrency Level:      100
Time taken for tests:   0.077 seconds
Complete requests:      10000
Failed requests:        0
Keep-Alive requests:    10000
Total transferred:      1470000 bytes
HTML transferred:       70000 bytes
Requests per second:    129676.46 [#/sec] (mean)
Time per request:       0.771 [ms] (mean)
Time per request:       0.008 [ms] (mean, across all concurrent requests)
Transfer rate:          18615.66 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.2      0       3
Processing:     0    1   0.3      1       3
Waiting:        0    1   0.3      1       3
Total:          0    1   0.4      1       5

Percentage of the requests served within a certain time (ms)
  50%      1
  66%      1
  75%      1
  80%      1
  90%      1
  95%      1
  98%      1
  99%      3
 100%      5 (longest request)

```
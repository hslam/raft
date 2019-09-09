

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
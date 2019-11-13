#!/bin/sh

nohup ./raftdb -h=localhost -p=7001 -c=8001 -f=9001 -d=6061 -m=8 -peers="" -path=./default.raftdb.1  >> default.out.1.log 2>&1 &
sleep 30s
curl -XPOST http://localhost:7001/db/foo -d 'bar'
sleep 1s
curl http://localhost:7001/db/foo
sleep 1s
./http_read_index -p=7001 -parallel=1 -total=100000 -clients=200 -bar=false
sleep 3s
./rpc_read_index -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=true -batch=true -batch_async=true -clients=8 -bar=false
sleep 3s
./http_write -p=7001 -parallel=1 -total=100000 -clients=200 -bar=false
sleep 3s
./rpc_write -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=true -batch=true -batch_async=true -clients=8 -bar=false
sleep 3s

killall raftdb

sleep 3s

nohup ./raftdb -h=localhost -p=7001 -c=8001 -f=9001 -d=6061 -m=8 -peers=localhost:9001,localhost:9002,localhost:9003 -path=./raftdb.1  >> out.1.log 2>&1 &
sleep 5s
nohup ./raftdb -h=localhost -p=7002 -c=8002 -f=9002 -d=6062 -m=8 -peers=localhost:9001,localhost:9002,localhost:9003 -path=./raftdb.2  >> out.2.log 2>&1 &
nohup ./raftdb -h=localhost -p=7003 -c=8003 -f=9003 -d=6063 -m=8 -peers=localhost:9001,localhost:9002,localhost:9003 -path=./raftdb.3  >> out.3.log 2>&1 &
sleep 30s
curl -XPOST http://localhost:7001/db/foo -d 'bar'
sleep 1s
curl http://localhost:7001/db/foo
sleep 1s
./http_read_index -p=7001 -parallel=1 -total=100000 -clients=200 -bar=false
sleep 3s
./rpc_read_index -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=true -batch=true -batch_async=true -clients=8 -bar=false
sleep 3s
./http_write -p=7001 -parallel=1 -total=100000 -clients=200 -bar=false
sleep 3s
./rpc_write -network=tcp -codec=pb -compress=gzip -h=127.0.0.1 -p=8001 -parallel=512 -total=100000 -multiplexing=true -batch=true -batch_async=true -clients=8 -bar=false
sleep 3s

killall raftdb

#!/bin/sh

ulimit -n 4096
nohup sh ./command.sh >> rpc.log.out 2>&1 &

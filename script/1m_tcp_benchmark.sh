#!/bin/bash

. ./script/env.sh

echo $line

. ./script/killall.sh

echo $line

. ./script/clean.sh

echo $line

. ./script/env.sh

echo $line

. ./script/build.sh

echo $line

. ./script/server.sh nbio_tcp

# echo $line

./script/tcpclient.sh -f=nbio_tcp -c=20000 -en=5000000 -b=1024 -rr=1

# echo $line
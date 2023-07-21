#!/bin/bash

. ./script/env.sh

echo "kill ${1} ..."
if type killall >/dev/null 2>&1; then
    killall -2 "${1}"
else
    pkill "${1}"
fi
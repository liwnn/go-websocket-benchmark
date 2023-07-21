#!/bin/bash

Connections=(10000)
BodySize=(1024)
BenchTime=(2000000)
SleepTime=5

frameworks=(
#    "fasthttp"
#    "gobwas"
#    "quickws"
#    "gorilla"
#    "gws"
#    "gws_std"
#    "hertz"
#    "hertz_std"
#    "nbio_blocking"
#    "nbio_mixed"
    "nbio_nonblocking"
#    "nbio_std"
#    "nettyws"
#    "nhooyr"
    "nbio_tcp"
)

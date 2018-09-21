#!/bin/sh
set -e

go build
docker build -t nuage/vnet-bgp-monitor:0.3 .

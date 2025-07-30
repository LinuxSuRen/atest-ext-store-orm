#!/bin/bash
set -e

SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
mkdir -p /root/.config/atest
mkdir -p /var/data

echo "start to run server"
nohup atest server&
cmd="atest run -p test-suite.yaml"

echo "start to run testing: $cmd"
kind=orm target=mysql:3306 driver=mysql $cmd

kind=orm target=mysql driver=mysql atest run -p testing-data-query.yaml
kind=orm target=greptimedb:4002 driver=greptime dbname=public atest run -p testing-data-query.yaml
kind=orm target=tdengine:6041 driver=tdengine password=taosdata dbname=information_schema atest run -p testing-data-query.yaml
kind=orm target=postgres driver=postgres atest run -p testing-data-query.yaml

kind=orm target=mysql driver=mysql atest run -p testing-data-query-mysql.yaml

cat /root/.config/atest/stores.yaml

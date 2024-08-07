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

cat /root/.config/atest/stores.yaml

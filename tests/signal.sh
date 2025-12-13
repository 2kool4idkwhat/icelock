#! /usr/bin/env bash

set -e

if [ "$1" = "unscoped" ]; then
  arg="--unscoped-ipc"
fi

sleep 5 &
test_pid=$!

icelock --rx /nix/store $arg -- kill $test_pid

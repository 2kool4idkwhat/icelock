#! /usr/bin/env bash

set -e

sleep 5 &
test_pid=$!

icelock --rx /nix/store -- kill $test_pid

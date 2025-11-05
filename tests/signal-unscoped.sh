#! /usr/bin/env bash

set -e

sleep 5 &
test_pid=$!

icelock --rx /nix/store --unscoped-ipc -- kill $test_pid

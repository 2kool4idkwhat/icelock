#! /usr/bin/env bash

icelock --rx /nix/store $@ -- keyctl list @us

#! /usr/bin/env bash

# NOTE: since this is an inherently impure test, we don't run it in the
# "basic" VM test script

url="$TEST_URL"

if [ "$url" = "" ]; then
  url="https://example.com"
fi

icelock --unrestricted-fs $@ -- curl --connect-timeout 5 $url

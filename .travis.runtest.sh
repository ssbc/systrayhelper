#!/bin/bash
set -e

if [[ $TRAVIS_OS_NAME != 'linux' ]]; then
  echo "integration testing is only supported on linux (using Xvfb and xdotool)"
  exit 1
fi

echo debug: $TRAVIS_OS_NAME $DISPLAY
go build -v -i
export PATH=$PATH:$(pwd) # expose helper

go test -v ./integration

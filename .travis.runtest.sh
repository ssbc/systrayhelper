#!/bin/bash
set -e

echo debug: $TRAVIS_OS_NAME $DISPLAY

if [[ $TRAVIS_OS_NAME == 'linux' ]]; then
  # "integration testing is only supported on linux (using Xvfb and xdotool)"
  go build -v -i
  export PATH=$PATH:$(pwd) # expose helper
  go test -v ./...
else
  go test -v
fi



#!/bin/bash

if [[ $TRAVIS_OS_NAME != 'linux' ]]; then
  echo "integration testing is only supported on linux (using Xvfb and xdotool)"
  exit 1
fi

echo debug: $TRAVIS_OS_NAME $DISPLAY
go build -v -i
export PATH=$PATH:$(pwd) // expose helper

go get -v github.com/stretchr/testify/require
go get -v github.com/cryptix/go/logging/logtest
go test ./integration

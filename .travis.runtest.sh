#!/bin/bash
set -e

echo debug: $TRAVIS_OS_NAME $DISPLAY

if [[ $TRAVIS_OS_NAME == 'linux' ]]; then
  # "integration testing is only supported on linux (using Xvfb and xdotool)"
  export PATH=$PATH:$(pwd) # expose helper
  command systrayhelper -v || {
    go build -v -i
    export PATH=$PATH:$(pwd)
  }
  xvfb-run -s '-screen 0 800x600x16' dbus-run-session go test -v ./...
else
  go test -v
fi



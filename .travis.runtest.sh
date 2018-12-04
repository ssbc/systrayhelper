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
  export TRAY_I3=t
  xvfb-run -e /tmp/xvfb.err dbus-run-session go test -timeout 2m -v ./...
  cat /tmp/xvfb.err || true
else
  go test -v
fi



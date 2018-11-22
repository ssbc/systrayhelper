#!/bin/bash
set -e

echo debug: $TRAVIS_OS_NAME $DISPLAY

if [[ $TRAVIS_OS_NAME == 'linux' ]]; then
  # "integration testing is only supported on linux (using Xvfb and xdotool)"
  go build -v -i
  export PATH=$PATH:$(pwd) # expose helper
  command systrayhelper -v || {
    go build -v
    export PATH=$PATH:$(pwd)
  }
  Xvfb ":23" -screen 0 800x600x16 &
  export DISPLAY=":23"
  export TRAY_XVFBRUNNING=t
  go test -v ./...
else
  go test -v
fi



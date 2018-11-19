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
  export TRAY_RECORD=1
  go test -v ./...
  # TODO: curl -H "Auhtorization:$SECRET_TOKEN1" --data-binary @"integration/test.mp4" https://boxbox
else
  go test -v
fi



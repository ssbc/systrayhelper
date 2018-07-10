#!/bin/bash

if [[ $TRAVIS_OS_NAME == 'osx' ]]; then
  brew install goreleaser/tap/goreleaser
fi

go get -t -v github.com/ssbc/systrayhelper/...

#!/bin/sh
set -e

TAR_FILE="/tmp/goreleaser.tar.gz"
RELEASES_URL="https://github.com/goreleaser/goreleaser/releases"
test -z "$TMPDIR" && TMPDIR="$(mktemp -d)"

last_version() {
  curl -sL -o /dev/null -w %{url_effective} "$RELEASES_URL/latest" | 
    rev | 
    cut -f1 -d'/'| 
    rev
}

download() {
  test -z "$VERSION" && VERSION="$(last_version)"
  test -z "$VERSION" && {
    echo "Unable to get goreleaser version." >&2
    exit 1
  }
  rm -f "$TAR_FILE"
  curl -s -L -o "$TAR_FILE" \
    "$RELEASES_URL/download/$VERSION/goreleaser_$(uname -s)_$(uname -m).tar.gz"
}

# nasty but I don't want to deal with argument jugling $@
if [[ $TRAVIS_OS_NAME == 'osx' ]]; then
  ln -s ./.goreleaser.darwin.yml ./.goreleaser.yml
elif [[ $TRAVIS_OS_NAME == 'linux' ]]; then
  ln -s ./.goreleaser.winAndLinux.yml ./.goreleaser.yml
fi

if command -v goreleaser; then
  goreleaser "$@"
else
  download
  tar -xf "$TAR_FILE" -C "$TMPDIR"
  "${TMPDIR}/goreleaser" "$@"
fi

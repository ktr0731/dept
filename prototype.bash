#!/bin/bash -e

if [[ "$#" -ne 1 ]]; then
  echo "usage: script <package>"
fi

if [[ ! -e "$GOPATH/src/$1" ]]; then
  echo "path: $GOPATH/src/$1 doesn't exist"
fi

if [[ ! -e "$GOPATH/src/$1/Gopkg.toml" ]]; then
  echo "Gopkg.toml doesn't exist"
fi

rm -rf ./tmp
mkdir -p ./tmp/src/github.com/ktr0731

echo "coping repo"
cp -r "$GOPATH/src/$1" ./tmp/src/github.com/ktr0731

OLD_GOPATH=$GOPATH
GOPATH=$PWD/tmp

echo $GOPATH

go install "$1"

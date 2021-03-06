#!/bin/sh
_cdir=$(cd $(dirname $0) && pwd)
cd -- "${_cdir}/.."
env GO111MODULE=on CGO_ENABLED=0 go build -tags netgo --ldflags '-s -w -extldflags "-static"'
env GO111MODULE=on GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags netgo --ldflags '-s -w -extldflags "-static"'

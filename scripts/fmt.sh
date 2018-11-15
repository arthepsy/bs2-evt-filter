#!/bin/sh
_cdir=$(cd $(dirname $0) && pwd)
cd -- "${_cdir}/.."
for i in . ./pkg/* ./internal/pkg/*; do
	cd -- "${_cdir}/.."
	cd -- "$i"
	go fmt
	go vet
done

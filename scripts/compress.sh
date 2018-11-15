#!/bin/sh
_cdir=$(cd $(dirname $0) && pwd)
cd -- "${_cdir}/.."
upx -9 bs2-evt-filter
upx -9 bs2-evt-filter.exe
exit 0

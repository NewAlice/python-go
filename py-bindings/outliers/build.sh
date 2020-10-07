#!/bin/bash
dp0="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
destdir=$dp0
#mkdir -p "$destdir"
#go get -u github.com/christian-korneck/go-python3 #force update
source "$dp0/set_env.sh"
#echo $PKG_CONFIG_PATH
go build -o "$destdir"


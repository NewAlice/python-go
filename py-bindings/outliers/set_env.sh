#!/bin/bash
dp0="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
py_version=385
export PKG_CONFIG_PATH=$dp0/pkg-config/py$py_version-$(uname -s)-$(uname -m)

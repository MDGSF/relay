#!/bin/bash

set -e

export GOOS=windows
export GOARCH=amd64

ScriptPath=$(cd `dirname $0` && pwd)

go build

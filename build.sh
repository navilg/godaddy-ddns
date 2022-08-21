#!/usr/bin/env bash

set -e

basedir=$(cd $(dirname $0) && pwd)

usage() {
    echo "bash build.sh "linux" "amd64" "1.0.0""
}

containerImageBuild() {
    docker build --build-arg OS=$1 --build-arg ARCH=$2 -t linuxshots/godaddy-ddns:$3 $basedir
}

binaryBuild() {
    go mod download
    GOOS=$1 GOARCH=$2 go build -o $basedir/assets/godaddyddns
}

############ MAIN #########
if [ $# -ne 3 ]; then
    echo "Exactly 3 arguments required."
    usage
    exit 1
fi
# binaryBuild $1 $2 $3
containerImageBuild $1 $2 $3
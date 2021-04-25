#!/bin/bash
set -e
set -x

DIR=`pwd`
NAME=`basename ${DIR}`
SHA=`git rev-parse --short HEAD`
VERSION=${VERSION:-$SHA}

GOOS=linux GOARCH=arm CGO_ENABLED=0 go build .
cp mikrotik-exporter dist/mikrotik-exporter_linux_arm

docker build -t jcfarsight/mikrotik-exporter:stable-armhf -f Dockerfile.armhf .
#docker push nshttpd/${NAME}:${VERSION}-armhf

rm mikrotik-exporter

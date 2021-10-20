#!/bin/bash
#

. $USER.env

VERSION=DEV

rm -rf dist 2>/dev/null
#go build -ldflags "-linkmode external -extldflags -static" -o ./dist/gpx_sensetif-datasource_linux_amd64 ./pkg
go build -o ./dist/gpx_sensetif-datasource_linux_amd64 ./pkg
yarn build || exit 1
mkdir -p $BUILD_DIR/sensetif-datasource 2>/dev/null
cp -r dist/* $BUILD_DIR/sensetif-datasource

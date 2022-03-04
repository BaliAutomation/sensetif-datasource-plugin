#!/bin/bash
#

. $USER.env

VERSION=DEV

rm -rf dist 2>/dev/null
#go build -ldflags "-linkmode external -extldflags -static" -o ./dist/gpx_sensetif-datasource_linux_amd64 ./pkg
go build -o ./dist/gpx_sensetif-datasource_linux_amd64 ./pkg
yarn build || exit 1

export GRAFANA_API_KEY=eyJrIjoiMzkwNTNkZTgxZTA4ODBjY2Q2YTIwNzg1NzBjZDAyOTNjOGNkZDU3OCIsIm4iOiJQdWJsaXNoIEtleSIsImlkIjo0OTA0MDZ9
npx @grafana/toolkit plugin:sign --rootUrls "https://sensetif.net/,http://localhost:3000/"

mkdir -p $BUILD_DIR/sensetif-datasource 2>/dev/null
cp -r dist/* $BUILD_DIR/sensetif-datasource

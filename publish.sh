#!/bin/bash
#

UNTRACKED=`git status | grep Untracked`
if [ "$UNTRACKED!" != "!" ] ; then
  echo "Repository is not committed."
  exit 1
fi

CHANGES=`git status | grep Changes`
if [ "$CHANGES!" != "!" ] ; then
  echo "Repository is not committed."
  exit 1
fi

VERSIONS=`git tag | grep "^[0-9]"`
VERSION=`echo "$VERSIONS" | sort -V | tail -1`

echo $VERSION

rm -rf dist 2>/dev/null
# go build -ldflags "-linkmode external -extldflags -static" -o ./dist/gpx_sensetif-datasource_linux_amd64 ./pkg
go build -o ./dist/gpx_sensetif-datasource_linux_amd64 ./pkg
yarn build || exit 1

export GRAFANA_API_KEY=eyJrIjoiMzkwNTNkZTgxZTA4ODBjY2Q2YTIwNzg1NzBjZDAyOTNjOGNkZDU3OCIsIm4iOiJQdWJsaXNoIEtleSIsImlkIjo0OTA0MDZ9
npx @grafana/toolkit plugin:sign --rootUrls "https://sensetif.net/,https://staging.sensetif.net/"

mkdir sensetif-datasource
cp -r dist/* sensetif-datasource
rm sensetif-app/module.js.LICENSE.txt 2>/dev/null
tar cfz sensetif-datasource_$VERSION.tar.gz sensetif-datasource
scp sensetif-datasource_$VERSION.tar.gz root@repo.sensetif.com:/var/www/repository/grafana-plugins/sensetif-datasource/
rm -rf sensetif-datasource sensetif-datasource_$VERSION.tar.gz

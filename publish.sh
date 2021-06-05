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
VERSION=`echo "$VERSIONS" | perl -Mversion -lane 'print join " ", sort { version->parse($a) cmp version->parse($b) } @F'  2>/dev/null | tail -1`

echo $VERSION

rm -rf dist 2>/dev/null
go build -o ./dist/gpx_sensetif-datasource-plugin_linux_amd64 ./pkg
yarn build || exit 1
mv dist sensetif-datasource-plugin
tar cf sensetif-datasource-plugin_$VERSION.tar.gz sensetif-datasource-plugin
scp sensetif-datasource-plugin_$VERSION.tar.gz root@repo.sensetif.com:/var/www/repository/grafana-plugins/sensetif-datasource-plugin/
rm -rf sensetif-datasource-plugin sensetif-datasource-plugin_$VERSION.tar.gz

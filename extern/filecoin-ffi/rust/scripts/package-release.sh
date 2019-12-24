#!/usr/bin/env bash

set -x

if [ -z "$1" ]; then
  TAR_FILE=`mktemp`.tar.gz
else
  TAR_FILE=$1
fi

TAR_PATH=`mktemp -d`

mkdir -p $TAR_PATH

find -L . -type f -name filecoin.h -exec cp -- "{}" $TAR_PATH/ \;
find -L . -type f -name libfilecoin.a -exec cp -- "{}" $TAR_PATH/ \;
find -L . -type f -name filecoin.pc -exec cp -- "{}" $TAR_PATH/ \;

pushd $TAR_PATH

tar -czf $TAR_FILE ./*

popd

rm -rf $TAR_PATH

echo $TAR_FILE

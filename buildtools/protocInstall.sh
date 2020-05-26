#!/bin/bash
set -e

if [ ! -d ./protoc ]; then
	VERSION=3.12.1
	ZIPNAME=protoc-$VERSION-osx-x86_64
	DOWNLOADLINK=https://github.com/protocolbuffers/protobuf/releases/download/v$VERSION/$ZIPNAME.zip
	curl -LO $DOWNLOADLINK
	unzip $ZIPNAME.zip -d protoc
	rm $ZIPNAME.zip
fi

if [ ! -d ./protoc-gen-go ]; then
	git clone --single-branch --depth 1 --branch "v1.4.2" https://github.com/golang/protobuf.git
	cd protobuf 
	go build -o ../protoc-gen-go/protoc-gen-go ./protoc-gen-go 
	cd ..
	rm -rf protobuf
fi

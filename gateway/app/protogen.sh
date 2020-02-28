#!/bin/bash

mkdir -p ./src/_proto

PROTOC=`command -v protoc`
if [[ "$PROTOC" == "" ]]; then
  echo "Required "protoc" to be installed. Please visit https://github.com/protocolbuffers/protobuf/releases (3.5.0 suggested)."
  exit -1
fi

echo "Compiling protobuf definitions"
protoc \
  --plugin=protoc-gen-ts=./node_modules/.bin/protoc-gen-ts \
  -I ../../index/ask/pb \
  -I ../../index/miner/pb \
  -I ../../index/slashing/pb \
  -I ../../deals/pb \
  -I ../../reputation/pb \
  -I ../../wallet/pb \
  -I ../../fpa/pb \
  --js_out=import_style=commonjs,binary:./src/_proto \
  --ts_out=service=grpc-web:./src/_proto \
  ../../index/ask/pb/ask.proto \
  ../../index/miner/pb/miner.proto \
  ../../index/slashing/pb/slashing.proto \
  ../../deals/pb/deals.proto \
  ../../reputation/pb/reputation.proto \
  ../../wallet/pb/wallet.proto \
  ../../fpa/pb/fpa.proto

for f in ./src/_proto/*.js
do
    echo '/* eslint-disable */' | cat - "${f}" > temp && mv temp "${f}"
done
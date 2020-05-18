LIB_VERSION=$1
PROTOS_PATH=$2
OUT_PATH=$3

LIB_PREFIX="@textile"
LIB_POSTFIX="grpc-powergate-client"
LIB_NAME="$LIB_PREFIX/$LIB_POSTFIX"
LIB_SRC_DIR="dist"

PROTOC_GEN_TS_PATH="$OUT_PATH/node_modules/.bin/protoc-gen-ts"
PTOTOC_GEN_OUT_DIR="$OUT_PATH/$LIB_SRC_DIR"

if [ -d "$OUT_PATH" ] 
then
  rm -rf $OUT_PATH
fi

mkdir -p $PTOTOC_GEN_OUT_DIR

printf '{"name":"%s", "version":"%s", "files":["%s"]}' $LIB_NAME $LIB_VERSION $LIB_SRC_DIR > $OUT_PATH/package.json

(cd $OUT_PATH && npm install --save-dev ts-protoc-gen)
(cd $OUT_PATH && npm install google-protobuf @improbable-eng/grpc-web)

ABS_OUT_PATH="$( cd $OUT_PATH >/dev/null 2>&1 ; pwd -P )"
ABS_PROTOS_PATH="$( cd $PROTOS_PATH >/dev/null 2>&1 ; pwd -P )"

PROTOS=$(find $ABS_PROTOS_PATH -path $ABS_OUT_PATH -prune -o -iname "*.proto" -print)

protoc \
    --plugin="protoc-gen-ts=${PROTOC_GEN_TS_PATH}" \
    --js_out="import_style=commonjs,binary:${PTOTOC_GEN_OUT_DIR}" \
    --ts_out="service=grpc-web:${PTOTOC_GEN_OUT_DIR}" \
    -I $ABS_PROTOS_PATH \
    ${PROTOS}
    

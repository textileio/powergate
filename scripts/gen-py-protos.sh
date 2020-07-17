LIB_VERSION=$1
PROTOS_PATH=$2
OUT_PATH=$3
SRC_PATH=$OUT_PATH/pow

if [ -d "$OUT_PATH" ] 
then
  rm -rf $OUT_PATH
fi

mkdir -p $SRC_PATH

cat << EOF > $OUT_PATH/setup.py
import setuptools

setuptools.setup(
    name="grpc-powergate-client",
    version="${LIB_VERSION}",
    author="Textile",
    author_email="contact@textile.io",
    url="https://github.com/textileio/powergate",
    packages=setuptools.find_packages(),
    install_requires=[
      'protobuf',
    ],
)
EOF

ABS_OUT_PATH="$( cd $OUT_PATH >/dev/null 2>&1 ; pwd -P )"
ABS_PROTOS_PATH="$( cd $PROTOS_PATH >/dev/null 2>&1 ; pwd -P )"

PROTOS=$(find $ABS_PROTOS_PATH -path $ABS_OUT_PATH -prune -o -iname "*.proto" -print)

python3 -m pip install grpcio-tools 
python3 -m grpc_tools.protoc -I$ABS_PROTOS_PATH --python_out=$SRC_PATH --grpc_python_out=$SRC_PATH $PROTOS

find $SRC_PATH -type d -exec touch {}/__init__.py \;

python3 -m pip install --user --upgrade setuptools wheel
cd $OUT_PATH && python3 setup.py sdist bdist_wheel


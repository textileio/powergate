LIB_VERSION=$1
PROTOS_PATH=$2
OUT_PATH=$3

if [ -d "$OUT_PATH" ] 
then
  rm -rf $OUT_PATH
fi

mkdir -p $OUT_PATH

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

./scripts/protoc_gen_plugin.bash --proto_path=$PROTOS_PATH --plugin_name=python --plugin_out=$OUT_PATH

find $OUT_PATH -type d -exec touch {}/__init__.py \;

python3 -m pip install --user --upgrade setuptools wheel
cd $OUT_PATH && python3 setup.py sdist bdist_wheel


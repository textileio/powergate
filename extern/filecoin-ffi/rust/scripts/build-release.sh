#!/usr/bin/env bash

set -Exeo pipefail

if [[ -z "$1" ]]
then
    (>&2 echo 'Error: script requires a library name, e.g. "filecoin" or "snark"')
    exit 1
fi

if [[ -z "$2" ]]
then
    (>&2 echo 'Error: script requires a toolchain, e.g. ./build-release.sh +nightly-2019-04-19')
    exit 1
fi

build_output_tmp=$(mktemp)

# clean up temp file on exit
#
trap '{ rm -f $build_output_tmp; }' EXIT

# build with RUSTFLAGS configured to output linker flags for native libs
#
RUSTFLAGS='--print native-static-libs' \
    cargo +$2 build \
    --release ${@:3} 2>&1 | tee ${build_output_tmp}

# parse build output for linker flags
#
linker_flags=$(cat ${build_output_tmp} \
    | grep native-static-libs\: \
    | head -n 1 \
    | cut -d ':' -f 3)

# generate pkg-config
#
sed -e "s;@VERSION@;$(git rev-parse HEAD);" \
    -e "s;@PRIVATE_LIBS@;${linker_flags};" "$1.pc.template" > "$1.pc"

# ensure header file was built
#
find -L . -type f -name "$1.h" | read

# ensure the archive file was built
#
find -L . -type f -name "lib$1.a" | read

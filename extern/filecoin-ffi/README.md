# Filecoin Proofs FFI

> C and CGO bindings for Filecoin's Rust libraries

## Building

To build and install libfilecoin, its header file and pkg-config manifest, run:

```shell
make
```

If no precompiled static library is available for your operating system, the
build tooling will attempt to compile a static library from local Rust sources.

### Forcing Local Build

To opt out of downloading precompiled assets, set `FFI_BUILD_FROM_SOURCE=1`:

```shell
rm .install-filecoin \
    ; make clean \
    ; FFI_BUILD_FROM_SOURCE=1 make
```

## License

MIT or Apache 2.0

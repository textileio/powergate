module github.com/textileio/filecoin

go 1.13

require (
	github.com/AndreasBriese/bbloom v0.0.0-20190823232136-616930265c33 // indirect
	github.com/filecoin-project/lotus v0.1.2-0.20191217122501-ae0864f8aba0
	github.com/golang/protobuf v1.3.2
	github.com/ipfs/go-cid v0.0.4
	github.com/ipfs/go-datastore v0.3.1
	github.com/ipfs/go-ds-badger v0.2.0 // indirect
	github.com/ipfs/go-log v1.0.0
	github.com/libp2p/go-libp2p v0.4.2 // indirect
	github.com/libp2p/go-libp2p-core v0.3.0
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.11 // indirect
	github.com/multiformats/go-varint v0.0.2 // indirect
	github.com/whyrusleeping/cbor-gen v0.0.0-20191216205031-b047b6acb3c0 // indirect
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413 // indirect
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553 // indirect
	golang.org/x/sys v0.0.0-20191210023423-ac6580df4449 // indirect
	google.golang.org/grpc v1.20.1
)

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi

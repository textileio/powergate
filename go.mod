module github.com/textileio/fil-tools

go 1.13

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/AlecAivazis/survey/v2 v2.0.6
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/GeertJohan/go.rice v1.0.0
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/caarlos0/spin v1.1.0
	github.com/containerd/continuity v0.0.0-20200107194136-26c1120b8d41 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/filecoin-project/go-address v0.0.0-20200107215422-da8eea2842b5
	github.com/filecoin-project/go-fil-markets v0.0.0-20200206024724-973498b060e3
	github.com/filecoin-project/go-sectorbuilder v0.0.2-0.20200203173614-42d67726bb62
	github.com/filecoin-project/lotus v0.2.8-0.20200212194405-7dc40c45168d
	github.com/gin-contrib/location v0.0.1
	github.com/gin-contrib/static v0.0.0-20191128031702-f81c604d8ac2
	github.com/gin-gonic/gin v1.5.0
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/golang/protobuf v1.3.3
	github.com/google/go-cmp v0.4.0
	github.com/google/uuid v1.1.1
	github.com/gorilla/websocket v1.4.1
	github.com/gosuri/uilive v0.0.4
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/improbable-eng/grpc-web v0.12.0
	github.com/ip2location/ip2location-go v8.3.0+incompatible
	github.com/ipfs/go-car v0.0.3-0.20200131220434-3f68f6ebd093
	github.com/ipfs/go-cid v0.0.5
	github.com/ipfs/go-datastore v0.4.4
	github.com/ipfs/go-ds-badger2 v0.0.0-20200123200730-d75eb2678a5d
	github.com/ipfs/go-ipfs-files v0.0.6
	github.com/ipfs/go-ipfs-http-client v0.0.6-0.20200205134739-3a5ff46efba6
	github.com/ipfs/go-ipld-cbor v0.0.5-0.20200204214505-252690b78669
	github.com/ipfs/go-log v1.0.2
	github.com/ipfs/go-log/v2 v2.0.2
	github.com/ipfs/interface-go-ipfs-core v0.2.6
	github.com/jessevdk/go-assets v0.0.0-20160921144138-4f4301a06e15
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/libp2p/go-libp2p v0.5.0
	github.com/libp2p/go-libp2p-core v0.3.1
	github.com/libp2p/go-libp2p-kad-dht v0.5.0
	github.com/logrusorgru/aurora v0.0.0-20200102142835-e9ef32dff381
	github.com/manifoldco/promptui v0.7.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/multiformats/go-multiaddr v0.2.0
	github.com/multiformats/go-multiaddr-dns v0.2.0
	github.com/multiformats/go-multihash v0.0.13
	github.com/olekukonko/tablewriter v0.0.4
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v0.1.1 // indirect
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.2
	github.com/stretchr/testify v1.5.1
	github.com/textileio/go-threads v0.1.9
	go.opencensus.io v0.22.3
	golang.org/x/crypto v0.0.0-20200117160349-530e935923ad // indirect
	golang.org/x/sys v0.0.0-20200121082415-34d275377bf9 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	google.golang.org/grpc v1.27.1
	gopkg.in/go-playground/validator.v9 v9.30.2 // indirect
	gopkg.in/yaml.v2 v2.2.7 // indirect
)

replace github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi

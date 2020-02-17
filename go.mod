module github.com/textileio/fil-tools

go 1.13

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/AlecAivazis/survey/v2 v2.0.5
	github.com/GeertJohan/go.rice v1.0.0
	github.com/caarlos0/spin v1.1.0
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/filecoin-project/go-address v0.0.0-20200107215422-da8eea2842b5
	github.com/filecoin-project/go-sectorbuilder v0.0.2-0.20200203173614-42d67726bb62
	github.com/filecoin-project/lotus v0.2.8-0.20200204184521-955b86deea71
	github.com/golang/protobuf v1.3.3
	github.com/google/go-cmp v0.4.0
	github.com/gorilla/websocket v1.4.1
	github.com/gosuri/uilive v0.0.4
	github.com/improbable-eng/grpc-web v0.12.0
	github.com/ip2location/ip2location-go v8.2.0+incompatible
	github.com/ipfs/go-cid v0.0.5
	github.com/ipfs/go-datastore v0.4.2
	github.com/ipfs/go-ds-badger2 v0.0.0-20200123200730-d75eb2678a5d
	github.com/ipfs/go-ipld-cbor v0.0.4
	github.com/ipfs/go-log v1.0.2
	github.com/ipfs/go-log/v2 v2.0.2
	github.com/libp2p/go-libp2p v0.5.0
	github.com/libp2p/go-libp2p-core v0.3.0
	github.com/libp2p/go-libp2p-kad-dht v0.5.0
	github.com/logrusorgru/aurora v0.0.0-20200102142835-e9ef32dff381
	github.com/manifoldco/promptui v0.7.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/multiformats/go-multiaddr v0.2.0
	github.com/multiformats/go-multihash v0.0.13
	github.com/olekukonko/tablewriter v0.0.4
	github.com/rs/cors v1.7.0 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.2
	github.com/stretchr/testify v1.4.0
	go.opencensus.io v0.22.3
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	google.golang.org/grpc v1.27.1
)

replace github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi

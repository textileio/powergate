module github.com/textileio/powergate/v2

go 1.16

require (
	contrib.go.opencensus.io/exporter/prometheus v0.2.0
	github.com/apoorvam/goterminal v0.0.0-20180523175556-614d345c47e5
	github.com/caarlos0/spin v1.1.0
	github.com/charmbracelet/bubbles v0.7.6
	github.com/charmbracelet/bubbletea v0.12.4
	github.com/containerd/continuity v0.0.0-20200228182428-0f16d7a0959c // indirect
	github.com/davidlazar/go-crypto v0.0.0-20200604182044-b73af7476f6c // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/filecoin-project/go-address v0.0.5
	github.com/filecoin-project/go-data-transfer v1.2.7
	github.com/filecoin-project/go-fil-markets v1.1.9
	github.com/filecoin-project/go-jsonrpc v0.1.4-0.20210217175800-45ea43ac2bec
	github.com/filecoin-project/go-state-types v0.1.0
	github.com/filecoin-project/lotus v1.5.0
	github.com/gin-contrib/location v0.0.2
	github.com/gin-contrib/static v0.0.0-20191128031702-f81c604d8ac2
	github.com/gin-gonic/gin v1.6.3
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.4
	github.com/google/uuid v1.2.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/improbable-eng/grpc-web v0.13.0
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-datastore v0.4.5
	github.com/ipfs/go-ds-badger2 v0.1.1-0.20200708190120-187fc06f714e
	github.com/ipfs/go-ipfs-files v0.0.8
	github.com/ipfs/go-ipfs-http-client v0.1.0
	github.com/ipfs/go-ipld-cbor v0.0.5
	github.com/ipfs/go-ipld-format v0.2.0
	github.com/ipfs/go-log/v2 v2.1.2-0.20200626104915-0016c0b4b3e4
	github.com/ipfs/interface-go-ipfs-core v0.4.0
	github.com/jessevdk/go-assets v0.0.0-20160921144138-4f4301a06e15
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/libp2p/go-libp2p v0.12.0
	github.com/libp2p/go-libp2p-core v0.7.0
	github.com/libp2p/go-libp2p-kad-dht v0.11.0
	github.com/logrusorgru/aurora v0.0.0-20200102142835-e9ef32dff381
	github.com/mattn/go-runewidth v0.0.10
	github.com/mitchellh/go-homedir v1.1.0
	github.com/muesli/termenv v0.7.4
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/multiformats/go-multiaddr-dns v0.2.0
	github.com/multiformats/go-multihash v0.0.14
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/olekukonko/tablewriter v0.0.5
	github.com/ory/dockertest/v3 v3.6.3
	github.com/oschwald/geoip2-golang v1.4.0
	github.com/otiai10/copy v1.4.2
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/textileio/dsutils v1.0.1
	github.com/textileio/go-ds-mongo v0.1.4
	go.opencensus.io v0.22.6
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	google.golang.org/grpc v1.35.0
	google.golang.org/protobuf v1.25.0
	honnef.co/go/tools v0.0.1-2020.1.3 // indirect
)

replace github.com/dgraph-io/badger/v2 => github.com/dgraph-io/badger/v2 v2.2007.2

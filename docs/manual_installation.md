# Powergate manual installation

This document describes the necessary steps for a manual installation setup. We highly recommend the dockerized installation version, which is in _.yaml_ files in the `docker` folder of the repo. Additionally, in that folder, we provide a _Makefile_ with one-command configuration to connect to different networks.

## Big picture

The following high-level design provides an excellent bird-eye picture of the different necessary components to run Powergate:
![Powergate Design](https://user-images.githubusercontent.com/6136245/86490337-6dbd5200-bd3d-11ea-9895-3689dd1c8a8f.png)

The relevant processes to understand from the picture above are:
- `powd`: The Powergate server daemon.
- `Lotus`: A Filecoin client.
- `go-ipfs`: An IPFS client.

The `powd` binary runs the Powergate daemon process, which provides all functionality of Powergate. To fully provide all its features, it requires two extra dependencies: `Lotus` and `go-ipfs`.

Below we'll provide the most natural order of installation of components.


## Install `go-ipfs`

To install [go-ipfs](https://github.com/ipfs/go-ipfs) please refer to its [installation instructions](https://github.com/ipfs/go-ipfs#install). Currently, Powergate targets version v0.6.1, so please have that consideration when selecting a version to install.

Powergate interacts with the IPFS daemon via its HTTP API, so there isn't a hard constraint on where the daemon should run. Depending on how decoupled you might want to do the setup, the daemon can live in the same or different host as `powd`.

The amount of resources that `go-ipfs` require will depend on the amount of load in the FFS-subsystem of Powergate, but take as a minimum the recommended [official system requirements](https://github.com/ipfs/go-ipfs#system-requirements). Recall those requirements are only for the `go-ipfs` daemon, so if you're installing other components in the same host, those are a hard-minimum.

Apart from CPU/Memory resources consideration, `go-ipfs` is the underlying system behind the _Hot Storage_ abstraction in the FFS module. That's to say, it's a datastore which should be fast to store and retrieve data. As a consequence, the amount of available storage for `go-ipfs` should be enough for your worst-case scenario estimation of _all_ the data stored in _Hot Storage_ for all FFS instances. Remember that storing data in `go-ipfs` has a slight overhead considering that the IPFS node does some data transformation. Despite that same transformation providing deduplication functionality that might compensate the overhead, it's recommended to plan for the worst use-case scenario with an extra threshold.

A priori, FFS instances can rely on the _Hot Storage_, thus the IPFS node, to go search for a particular piece of data in the public IPFS network which means that the IPFS daemon should have internet connection, and enough bandwidth to satisfy your use case.

We're currently not targeting any particular `go-ipfs` extra configuration such as custom networking configurations, or specific _profiles_. This might change in the future when we have more experience with better setups for `go-ipfs` in production and high load environments. Additionally, we'll soon explore more reliable and scalable setups for the _Hot Storage_ such as IPFS cluster, so stay tuned for more information.


## Install a Filecoin client (`lotus`)

Powergate interacts with the Filecoin network using two mechanisms:
1. It runs a libp2p client that connects to the Filecoin network, and it's DHT. This client is only used by the indices to interact directly with miners to resolve geolocation information, liveness, and other features.
2. It interacts with the Filecoin blockchain using a client. Currently, only `lotus` is supported since is the only fully-featured and most stable Filecoin client.

To install `lotus` there are two different flavors:
- Raw binary run: To run `lotus` daemon-less, please refer to the [official documentation](https://docs.lotu.sh/en+install-lotus-ubuntu) which provides all steps necessary depending on the OS of the host. 
- Run with _systemd_: You can run Lotus as a service with systemd, if that's the case refer to [this](https://docs.lotu.sh/en+install-systemd-services) link.

One important fact about `lotus` for Powergate, is that it only needs to be run in _client mode_. The `lotus` node *wont't* be a miner, thus it doesn't need a _Lotus storage miner_, workers, or similar components.

The official system requirements for running Lotus can be found [here](https://docs.lotu.sh/en+faqs#what-operating-systems-can-lotus-run-on-115423).

### Lotus configuration

Apart from installing a pristine Lotus daemon, some further configuration should be provided to work correctly with Powergate. 

By default, the `lotus` daemon creates a folder in `~/.lotus` containing all the daemon state. For configuring Lotus, there are two options:
- Editing `~/.lotus/config.toml` which contains the configuration parameters in TOML format.
- Using environment variables with the `LOTUS_` prefix, which takes precedence over the file configuration.

The default Lotus configuration is similar to the following snippet:
```
# Default config:
[API]
#  ListenAddress = "/ip4/127.0.0.1/tcp/1234/http"
#  RemoteListenAddress = ""
#  Timeout = "30s"
#
[Libp2p]
#  ListenAddresses = ["/ip4/0.0.0.0/tcp/0", "/ip6/::/tcp/0"]
#  AnnounceAddresses = []
#  NoAnnounceAddresses = []
#  ConnMgrLow = 150
#  ConnMgrHigh = 180
#  ConnMgrGrace = "20s"
#
[Pubsub]
#  Bootstrapper = false
#  RemoteTracer = "/ip4/147.75.67.199/tcp/4001/p2p/QmTd6UvR47vUidRNZ1ZKXHrAFhqTJAD27rKL9XYghEKgKX"
#
[Client]
#  UseIpfs = false
#  IpfsMAddr = ""
#  IpfsUseForRetrieval = false
#
[Metrics]
#  Nickname = ""
#  HeadNotifs = false
```

The first necessary configuration is allowing Powergate to access the [Lotus JSON-RPC API](https://docs.lotu.sh/en+api). By default, the Lotus daemon listens on `127.0.0.1:1234`. In case of needing to change the listening host/port, the configuration `API.ListenAddress` or `LOTUS_API_LISTENADDRESS` should be changed to the corresponding [multiaddress](https://multiformats.io/multiaddr/).

The second necessary configuration, is leveraging integrating the IPFS node, `go-ipfs` mentioned in the previous section, with Lotus. This configuration allows the IPFS node to be used as the underlying blockstore to store and retrieva data from the Filecoin network. This provides a convenient and efficient data-flow path in which Powergate won't participate, and thus avoid incurring in more overhead and resource usage.

To integrate IPFS with the Lotus node, the following configuration settings should be set:
- `Client.UseIpfs`/`LOTUS_CLIENT_USEIPFS`: With value `true`. This indicates that Lotus will use IPFS as the underlying blockstore for data for storage deals.
- `Client.IpfsUseForRetrieval`/`LOTUS_CLIENT_IPFSUSEFORERETRIEVAL`: With value `true`. This indicates Lotus will use IPFS as the underlying blockstore for retrieval deals.
- `Client.IpfsMAddr`/`LOTUS_CLIENT_IPFSMADDR`: Should contain the multiaddress of the IPFS node API mentioned in the last section, e.g: `/ip4/192.168.11.12/tcp/5001`.

The above description wraps all necessary and minimal configuration to run Lotus for Powergate.

### Wallet address keys

As mentioned before, the default folder where Lotus will save all its state is `~/.lotus`. Apart from containing the `config.toml` file, it includes the `keystore` folder. This folder contains *all the keys of wallets created in Lotus*. It's essential to understand the consequences of this fact correctly. 

The Lotus node will manage all wallet addresses used by Powergate, thus its private keys are contained in this folder. Therefore, take necessary precautions such as:
- This folder should have minimal access.
- Frequent backups are recommended for this folder. Recall that keys-backups are sensitive process to handle correctly since you're making copies of keys, and possibly increasing the risk of leakeage.
- You can reduce the blast-radius of data loss by thinking about different strategies to keep the balances from these wallet addresses to a minimum.

There's some discussion to allow Lotus to use an external component for signing, which allows to minimize these risks. Whenever that feature is ready, there's a high chance that this document will be updated to explain how to leverage that new feature for Powergate.


# Install Powergate

To run Powergate, you should run the `powd` binary with proper flags or environment variables configuration.

You can download `powd` binaries from the [GitHub Releases section](https://github.com/textileio/powergate/releases) for your OS and architecture. If you prefer to compile `powd` from source, you can checkout the Powergate repo to your desired version and run `make install-powd` which will build and install `powd` in your `$GOPATH/bin` folder. Note that to compile, we're targeting Go 1.14 or newer version.

## Basic configuration

In this section, we outline the basic configuration needed for Powergate. Recall you can execute `powd -h` to look for default values and format of configuration values.

The first configuration step is to provide information to connect to the Lotus API correctly. For this, you should provide `POWD_LOTUSHOST`/`--lotushost`, which should be multiaddress that indicates where is the Lotus API JSON-RPC endpoint mentioned in the previous section. 

Additionally, you should indicate which is the _auth token_ of the API. The _auth token_ lives in `~/.lotus/token` in your Lotus host. Powergate allows this parameter to be configured in two ways: the path of this token file, or providing the token value directly. For the former, you should set `POWD_LOTUSTOKENFILE`/`--lotustokenfile`, and for the latter `POWD_LOTUSTOKEN`/`--lotustoken`. At least one of each env/flags should be provided. If that isn't the case `powd` will fail to start indicating that as an error.

The second configuration step is to provide information to connect to the IPFS node. This is done by configuring `POWD_IPFSAPIADDR`/`--ipfsapiaddr` which is a multiaddress with the `go-ipfs` API endpoint.

Finally, the last necessary step is to have the _Geolite database_ which provides geolocation information for Powergate indices. This database can also be found in the release assets with name `GeoLite2-City.mmdb`. By default, `powd` expect to find this file in the current executing folder. You can customize the path of this file with `POWD_MAXMINDDBFOLDER`/`--maxminddbfolder`.

## gRPC APIs

Powergate provides a gRPC endpoint to serve its APIs, and a [grpc-Web proxy](https://github.com/improbable-eng/grpc-web) for the JS client. The default listening address can be changed by modifying `POWD_GRPCHOSTADDR`/`--grpchostaddr` and `POWD_GRPCWEBPROXYADDR`/`--grpchostaddr` respectively.

The `powd` binary doesn't have flags for providing SSL certificates to serve these endpoints securely. You might consider using a reverse proxy which does SSL offloading. The gRPC clients support secure gRPC connections.

## Filecoin network selection

Although Powergate interacts with the Filecoin network through Lotus, it also connects to the Filecoin network and DHT directly. This is done to gather information for building indices, which means that is important that Powergate connects to the same Filecoin network as the Lotus node to produce coherent index data.

To avoid possible configuration mismatches, Powergate always connects to the same network Lotus is connected. This is done automatically by Powergate which asks the Lotus node for the network identity its connected to. No extra flags or configuration are needed to select the right network to connect to. 

## Indices gateway

Powergate provides an indices gateway which presents miners, storage asks, and faults indices in a pretty web page.x  The listening address of this webserver can be modified with `POWD_GATEWAYHOSTADDR`/`--gatewayhostaddr`.

## FFS configuration

The FFS module is a central part of Powergate for storing data in Filecoin in a declarative way.

To provide some context about the related configuration parameters, we should understand some basics about FFS. In the FFS module, we have two central actors: the _FFS manager_, and _FFS instances_. The _FFS Manager_ is responsible for creating and managing _FFS instances_, which provide a scoped container for storing data in Filecoin.

In the usual configuration of FFS, when the manager creates a new instance it also creates a new wallet address in Lotus to be owned by it. 

As an optional feature, it can fund the newly created address with initial funds. For this to happen, the manager should have a _masteraddr_. A _masteraddr_ is a wallet address owned by the manager, which is used as the source to fund newly created instances wallet addresses. 

For this feature to be enabled, the `POWD_LOTUSMASTERADDR`/`--lotusmasteraddr` should be set. If this is the case, this most probably involved a previous action from the system administrator to create or import the wallet address with funds in the Lotus node. If you're running Powergate in a test network, you can avoid setting a specific _master address_, and enable the `POWD_AUTOCREATEMASTERADDR`/`--autocreatemasteraddr`. This configuration will auto-create a _master address_ for the manager, and auto-fund it with the corresponding test network facuet.

Finally, to configure the amount of _attoFil_ to be transferred by the _manager_ to a newly created wallet, you should set `POWD_WALLETINITIALFUND`/`--walletinitialfund`. This configuration only applies if you set the _master address_ manually, or enabled the master address auto-creation.

### Sharing wallet address in all FFS instances

Depending on your use-case, having each FFS instance have its own wallet address might be a beneficial setup. If you plan different FFS instances for different purposes, it may be logical to bound the amount of FIL spent on storage or retrieval actions for various purposes. Also, any abuse of FIL spending gets a hard limit per-instance and not in a _single bag_ of FIL.

On the other hand, other use-cases might want to share a single wallet address for all FFS instances operation. An example could be saving fees or gas cost of funding transactions, or avoiding the delay of funding transactions to be accepted on-chain.

If that's the case, you can provide an extra configuration `POWD_FFSUSEMASTERADDR`/`--ffsusemasteraddr` which indicates to the manager to avoid creating a new wallet address when an instance is created. As a consequence, _all newly created FFS instances will use the manager master address_ as its funding wallet address for operations.

## Metrics endpoint

Powergate exposes an Prometheus metrics endpoint at `:8888`. You can connect your Prometheus scrapper to gather metrics data about multiple components of Powergate. In the `docker/grafana` folder you can find a possible Grafana dashboard to leverage using existing metrics.

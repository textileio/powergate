# Powergate

[![Made by Textile](https://img.shields.io/badge/made%20by-Textile-informational.svg?style=popout-square)](https://textile.io)
[![Chat on Slack](https://img.shields.io/badge/slack-slack.textile.io-informational.svg?style=popout-square)](https://slack.textile.io)
[![GitHub license](https://img.shields.io/github/license/textileio/filecoin.svg?style=popout-square)](./LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/textileio/powergate?style=flat-square)](https://goreportcard.com/report/github.com/textileio/powergate?style=flat-square)
[![GitHub action](https://github.com/textileio/powergate/workflows/Tests/badge.svg?style=popout-square)](https://github.com/textileio/powergate/actions)

Powergate is a multitiered file storage API built on Filecoin and IPFS, and and index builder for Filecoin data. It's designed to be modular and extensible.

Join us on our [public Slack channel](https://slack.textile.io/) for news, discussions, and status updates. [Check out our blog](https://medium.com/textileio) for the latest posts and announcements.

*Warning* This project is still **pre-release** and is not ready for production usage.

## Table of Contents

-   [Design](#design)
-   [API + s CLI](#api&CLI)
-   [Installation](#installation)
-   [Contributing](#contributing)
-   [Changelog](#changelog)
-   [License](#license)

## Design

Powergate is composed of different modules which can be used independently, and compose toegheter to provide other higher-level modules.

### 📢 Deals module
The Deals module provides a lower layer of abstraction to a Filecoin client node. It provides simple APIs to store, watch, and retrieve data in the Filecoin network. Currently, it interacts with the Lotus client but we have plans to support other Filecoin clients.

### 👷 Indices and Reputation scoring
Powergate builds three indexes related to on-chain and off-chain data.

The _Miners index_ provides processed data regarding registered miners (on-chain and off-chain), such as: total miner power, relative power, online status, geolocation, and more!

The _Ask index_ provides a fast-retrieval up to date snapshot of miner's asking prices for data storage.

The _Slashing index_ provides history data about miners faults while proving their storage on-chain. 

Built on top of the previous indexes, a _Reputation_ module constructs a weighted-scoring system that allows to sort miners considering multiple on-chain and off-chain data, such as: compared price to the median of the market, low storage-fault history, power on network, and external sources (soon!).

###  ⚡ FFS
This module provides a multitiered file storage API built on Filecoin and IPFS. Storing data on IPFS and Filecoin is as easy as expressing your desired configuration for storing a Cid.

Want to know more about this Powergate module? Check out our presentation and demo at the _IPFS Pinning Summit_:
[![Video](https://img.youtube.com/vi/aiOTSkz_6aY/0.jpg)](https://youtu.be/aiOTSkz_6aY)

### 💫 API + CLI

Powergate expose modules functionalities through gRPC endpoints. 
You can explore our `.proto` files to generate your clients, or take advange of a ready-to-use Powergate Go and [JS client](https://github.com/textileio/js-powergate-client). 🙌

We have a CLI that supports most of Powergate features.
```bash
$ make build-cli
$ ./pow --help
A client for storage and retreival of powergate data

Usage:
  pow [command]

Available Commands:
  asks        Provides commands to view asks data
  deal        Interactive storage deal
  deals       Provides commands to manage storage deals
  ffs         Provides commands to manage ffs
  health      Display the node health status
  help        Help about any command
  init        Initializes a config file with the provided values or defaults
  miners      Provides commands to view miners data
  net         Provides commands related to peers and network
  reputation  Provides commands to view miner reputation data
  slashing    Provides commands to view slashing data
  wallet      Provides commands about filecoin wallets

Flags:
      --config string          config file (default is $HOME/.powergate.yaml)
  -h, --help                   help for pow
      --serverAddress string   address of the powergate service api (default "/ip4/127.0.0.1/tcp/5002")

Use "pow [command] --help" for more information about a command.
```

## Installation

Powergate installation involves getting running external dependencies, and wiring them correctly with Powergate.

### External dependencies
Powergate needs external dependencies in order to provide full functionality, in particular a synced Filecoin client and a IPFS node.


#### Filecoin client
Currently we support the Lotus Filecoin client, but we plan to support other clients.

All described modules of Powergate need to comunicate with Lotus to build indeces data, and provide storing and retrieving features in FFS. To install Lotus refer to its [official](https://lotu.sh/) documentation, taking special attention to [its dependencies](https://docs.lotu.sh/en+install-lotus-ubuntu). 

Fully syncing a Lotus node can take time, so be sure to check you're fully synced doing `./lotus sync status`.

We also automatically generate a public Docker image targeting the `master` branch of Lotus. This image is a pristine version of Lotus, with a sidecar reverse proxy to provide external access to the containerized API. For more information, refer to [textileio/lotus-build](https://github.com/textileio/lotus-build) and its [Dockerhub repository](https://hub.docker.com/repository/docker/textile/lotus).

### IPFS node
You should be running an IPFS node. You can refer [here](https://docs.ipfs.io/guides/guides/install/) for installation instructions to run native binaries or the [Dockerhub repository](https://hub.docker.com/r/ipfs/go-ipfs) if you want to run a contanerized version.

### Server
To build the Powergate server, run:
```bash
make build-server
```
You can run the `-h` flag to see the configurable flags:
```bash
$ ./powd -h
Usage of ./powd:
      --debug                     enable debug log levels
      --embedded                  run in embedded ephemeral FIL network
      --gatewayhostaddr string    gateway host listening address (default "0.0.0.0:7000")
      --grpchostaddr string       grpc host listening address (default "/ip4/0.0.0.0/tcp/5002")
      --grpcwebproxyaddr string   grpc webproxy listening address (default "0.0.0.0:6002")
      --ipfsapiaddr string        ipfs api multiaddr (default "/ip4/127.0.0.1/tcp/5001")
      --lotushost string          lotus full-node address (default "/ip4/127.0.0.1/tcp/1234")
      --lotusmasteraddr string    lotus addr to be considered master for ffs
      --lotustoken string         lotus full-node auth token
      --lotustokenfile string     lotus full-node auth token file
      --pprof                     enable pprof endpoint
      --repopath string           repo-path (default "~/.powergate")
      --walletinitialfund int     created wallets initial fund in attoFIL (default 4000000000000000)
pflag: help requested
```

We'll soon provide better information about Powergate configurations, stay tuned! 📻

### Run in _Embedded mode_

The server can run in _Embedded_ mode which auto-creates a fake devnet with a single miner and connects to it.
The simplest way to run it is:
```bash
cd docker
make embed
```

This creates an ephemeral server with all working for CLI interaction.


Here is a full example of using the embedded network:

Terminal 1:
```bash
cd docker
make embed
```
Wait for seeing logs about the height of the chain increase in a regular cadence.

Terminal 2:
```bash
make build
❯ head -c 700 </dev/urandom > myfile
❯ ./pow ffs create
> Instance created with id 0ac0fb4d-581c-4276-bd90-a9aa30dd4cb4 and token 883f57b1-4e66-47f8-b291-7cf8b10f6370
❯ ./pow ffs addToHot -t 883f57b1-4e66-47f8-b291-7cf8b10f6370 myfile
> Success! Added file to FFS hot storage with cid: QmYaAK8SSsKJsJdtahCbUe7MZzQdkPBybFCcQJJ3dKZpfm
❯ ./pow ffs push -w -t 883f57b1-4e66-47f8-b291-7cf8b10f6370 QmYaAK8SSsKJsJdtahCbUe7MZzQdkPBybFCcQJJ3dKZpfm
> Success! Pushed cid config for QmYaAK8SSsKJsJdtahCbUe7MZzQdkPBybFCcQJJ3dKZpfm to FFS with job id: 966dcb44-9ef4-4d2a-9c90-a8103c77c354
               JOB ID                   STATUS
966dcb44-9ef4-4d2a-9c90-a8103c77c354    Success
❯ ./pow ffs get  -t 883f57b1-4e66-47f8-b291-7cf8b10f6370 QmYaAK8SSsKJsJdtahCbUe7MZzQdkPBybFCcQJJ3dKZpfm myfile2
> Success! Data written to myfile2
```

Notes:
- A random `myfile` is a small random file since the devnet is running with a constrained sectorbuilder mock and sector size. Sizes close to ~700 bytes should be fine.
- The devnet might run correctly for 150 epochs before it can become unstable.


### Run in full mode

Running the _full mode_ can be done by:
```bash
cd docker
make up
```

This will spinup and auto-wire:
- _Prometheus_ ,endpoint for metrics
- _Grafana_, for metrics dashboard
- _cAdvisor_, for container metrics
- _Lotus_ node configured for testnet.
- _Powergate_, wired with all of above components.

Recall that you should wait for _Lotus_ to be fully-synced which might take a long time now.
If you're running the _Lotus_ node in the host and want to leverage its fully synced, you could:
- Bind the `.lotus` folder to the _Lotus_ node in the docker-compose file.
- Or, just `docker cp` it.
In any option, you should stop the original _Lotus_ node.

If you don't have a fully-synced _Lotus_ node and don't want to wait, consider using our [archives](https://lotus-archives.textile.io/).

## Test
For running tests: `make test`

## Benchmark
There's a dedicated binary to run benchmarks against a Powergate server. 
For more information see the [specific README](exe/bench/README.md).


## Docker

A `powd` Docker image is available at [textile/powergate](https://hub.docker.com/r/textile/powergate) on DockerHub.

## Contributing

This project is a work in progress. As such, there's a few things you can do right now to help out:

-   **Ask questions**! We'll try to help. Be sure to drop a note (on the above issue) if there is anything you'd like to work on and we'll update the issue to let others know. Also [get in touch](https://slack.textile.io) on Slack.
-   **Open issues**, [file issues](https://github.com/textileio/powergate/issues), submit pull requests!
-   **Perform code reviews**. More eyes will help a) speed the project along b) ensure quality and c) reduce possible future bugs.
-   **Take a look at the code**. Contributions here that would be most helpful are **top-level comments** about how it should look based on your understanding. Again, the more eyes the better.
-   **Add tests**. There can never be enough tests.

Before you get started, be sure to read our [contributors guide](./CONTRIBUTING.md) and our [contributor covenant code of conduct](./CODE_OF_CONDUCT.md).

## Changelog

[Changelog is published to Releases.](https://github.com/textileio/powergate/releases)

## License

[MIT](LICENSE)

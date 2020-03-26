# Powergate

[![Made by Textile](https://img.shields.io/badge/made%20by-Textile-informational.svg?style=popout-square)](https://textile.io)
[![Chat on Slack](https://img.shields.io/badge/slack-slack.textile.io-informational.svg?style=popout-square)](https://slack.textile.io)
[![GitHub license](https://img.shields.io/github/license/textileio/filecoin.svg?style=popout-square)](./LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/textileio/powergate?style=flat-square)](https://goreportcard.com/report/github.com/textileio/powergate?style=flat-square)
[![GitHub action](https://github.com/textileio/powergate/workflows/Tests/badge.svg?style=popout-square)](https://github.com/textileio/powergate/actions)

> Textile's Filecoin swiss army knife for developers

Join us on our [public Slack channel](https://slack.textile.io/) for news, discussions, and status updates. [Check out our blog](https://medium.com/textileio) for the latest posts and announcements.

## Table of Contents

-   [Usage](#usage)
-   [Contributing](#contributing)
-   [Changelog](#changelog)
-   [License](#license)

## Usage

*Warning* This project is still **pre-release** and is only meant for testing.

### Lotus (`lotus`)

_Powergate_ communicates with a _Lotus_ node to interact with the filecoin network.
If you want to run _Powergate_ targeting the current testnet, you should be running a fully-synced Lotus node in the same host as _Powergate_.
For steps to install _Lotus_, refer to  [https://lotu.sh/](https://lotu.sh/) taking special attention to [its dependencies](https://docs.lotu.sh/en+install-lotus-ubuntu). 

Since bootstrapping a _Lotus_ node from scratch and getting it synced may take too long, _Powergate_ allows an `--embedded` flag, which 
auto-creates a fake local testnet with a single miner, and auto-connects to it. This means, only running the _Powergate_ server with the flag enabled, allows to use it in some reasonable context with almost no extra setup.

For both building the _CLI_ and the _Server_, run:
```bash
make build
```
This will create `powd` (server) and `pow` (CLI) of _Powergate_.

### Client (`pow`)

To build the CLI, run:
```bash
make build-cli
```

Try `pow --help`.

### Server 

To build the Server, run:
```bash
make build-server
```

The server connects to _Lotus_ and enables multiple modules, such as:
- Reputations module:
   - Miners index: with on-chain and metadata information.
   - Ask index: with an up-to-date information about available _Storage Ask_ in the network.
   - Slashing index: contains a history of all miner-slashes.
- Deals Module:
    - Contain helper features for making and watching deals.
- FFS: 
    - A powerful level of abstraction to pin Cids in Hot and Cold storages, more details soon!


### Run in _Embedded mode_

The server can run in _Embedded_ mode which auto-creates a fake devnet with a single miner and connects to it.
The simplest way to run it is:
```bash
cd docker
make embed
```

This creates an ephemeral server with all working for CLI interaction.

### Run in full mode

Running the _full mode_ can be done by:
```bash
cd docker
make fresh
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

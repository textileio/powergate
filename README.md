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

See [https://lotu.sh/](https://lotu.sh/). Required for client. Lotus is an implementation of the Filecoin Distributed Storage Network—we run the Lotus client to join the Filecoin Testnet. 

### Client (`pow`)

```bash
go build -i -o pow exe/cli/main.go 
chmod +x pow 
```

Try `pow --help`.

### Server 
The server connects to Lotus and enables multiple modules, such as:
- Reputations module:
   - Miners index: with on-chain and metadata information.
   - Ask index: with an up-to-date information about available _Storage Ask_ in the network.
   - Slashing index: contains a history of all miner-slashes.
- Deals Module:
    - Contain helper features for making and watching deals.
### Prerequisites
Currently, the server needs the same dependencies as Lotus, so most probably they will be already installed if you run the server in _dev mode_ (`go run` syntax). If that's not the case, [see here](https://docs.lotu.sh/en+install-lotus-ubuntu).

### Run in _dev mode_
The server can be run in _dev mode_ with `go run exe/server/main.go`. 

### Run in _docker mode_
You can run the server in a _Docker Compose_ enviroment with _Prometheus_ and _Grafana_ to have a health-monitoring for the server.
To do so:
```
cd docker
make fresh
```
This will do whatever it takes to get all systems running:
- `localhost:3000`: Grafana dashboard with anonymous setup. _admin_:_foobar_ are  admin credentials.
- `localhost:9090`: Prometheus enpoint.
- `localhost:8888`: Server Prometheus endpoint for metric scraping.
- `localhost:8889`: HTTP endpoint for current index values (`index/miners`, `index/slashing`, `index/ask`).

It's important to mention that the _Docker Compose_ is run in _host_ _network mode_, which is only supported in Linux. Eventually we'll include Lotus inside the compose file allowing for default network modes.

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

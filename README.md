# filecoin

[![Made by Textile](https://img.shields.io/badge/made%20by-Textile-informational.svg?style=popout-square)](https://textile.io)
[![Chat on Slack](https://img.shields.io/badge/slack-slack.textile.io-informational.svg?style=popout-square)](https://slack.textile.io)
[![GitHub license](https://img.shields.io/github/license/textileio/filecoin.svg?style=popout-square)](./LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/textileio/filecoin?style=flat-square)](https://goreportcard.com/report/github.com/textileio/filecoin?style=flat-square)
[![GitHub action](https://github.com/textileio/filecoin/workflows/Tests/badge.svg?style=popout-square)](https://github.com/textileio/filecoin/actions)

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

### Client (`filcoin`)

    go build -i -o filecoin exe/cli/main.go 
    chmod +x filecoin 

Try `filecoin --help`.

## Contributing

This project is a work in progress. As such, there's a few things you can do right now to help out:

-   **Ask questions**! We'll try to help. Be sure to drop a note (on the above issue) if there is anything you'd like to work on and we'll update the issue to let others know. Also [get in touch](https://slack.textile.io) on Slack.
-   **Open issues**, [file issues](https://github.com/textileio/filecoin/issues), submit pull requests!
-   **Perform code reviews**. More eyes will help a) speed the project along b) ensure quality and c) reduce possible future bugs.
-   **Take a look at the code**. Contributions here that would be most helpful are **top-level comments** about how it should look based on your understanding. Again, the more eyes the better.
-   **Add tests**. There can never be enough tests.

Before you get started, be sure to read our [contributors guide](./CONTRIBUTING.md) and our [contributor covenant code of conduct](./CODE_OF_CONDUCT.md).

## Changelog

[Changelog is published to Releases.](https://github.com/textileio/filecoin/releases)

## License

[MIT](LICENSE)

# Wibble

![Captain Blackadder](./docs/wibble.jpg)

A simple TUI for RSS Feeds. It has the following features:

- Register feeds via a configuration file.
- Mark articles as read.
- Open an article in the browser.
- Open a modal view of the article summary in the application.
- Override all the colours used via the configuration file.
- Refresh feeds every 10 minutes (configurable).

Rose Pine Day

![Rose Pine Day](./docs/rose-pine-day.png)

Rose Pine Night

![Rose Pine Night](./docs/rose-pine-night.png)

You can theme it however you like, see [docs/examples/](./docs/examples/) for other themes.

## Requirements

- Go 1.26+

## Installation

Clone the repository and build the CLI:

```shell
git clone https://github.com/benmatselby/wibble.git
cd wibble
make build
```

## Usage

Run the CLI:

```shell
./wibble --help
```

The application relies on a configuration file. Please see examples in `docs/examples/`. Running `./wibble --help` will explain where the configuration file is expected to be found. You can also specify a custom path to the configuration file using the `--config` flag.

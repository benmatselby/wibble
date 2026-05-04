# Wibble

![Captain Blackadder](./docs/wibble.jpg)

A simple TUI for RSS Feeds. It has the following features:

- Register feeds via a configuration file.
- Mark articles as read.
- Override all the colours used via the configuration file. See (`examples`)
- Refresh feeds every 10 minutes (configurable).

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

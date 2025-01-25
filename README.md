# pplx

[![GitHub release](https://img.shields.io/github/release/sgaunet/pplx.svg)](https://github.com/sgaunet/pplx/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/sgaunet/pplx)](https://goreportcard.com/report/github.com/sgaunet/pplx)
![GitHub Downloads](https://img.shields.io/github/downloads/sgaunet/pplx/total)
[![GoDoc](https://godoc.org/github.com/sgaunet/pplx?status.svg)](https://godoc.org/github.com/sgaunet/pplx)
[![License](https://img.shields.io/github/license/sgaunet/pplx.svg)](LICENSE)

It's an unofficial CLI program to query/chat with the [perplexity API](https://www.perplexity.ai/).

## Installation

## Option 1

* Download the latest release from the [releases page](https://github.com/sgaunet/pplx/releases).
* Install the binary in /usr/local/bin or any other directory in your PATH.

## Option 2: With brew

```sh
brew tap sgaunet/homebrew-tools
brew install pplx
```

## Usage

```sh
Program to interact with the Perplexity API.

        You can use it to chat with the AI or to query it.

Usage:
  pplx [command]

Available Commands:
  chat        chat subcommand is an interactive chat with the Perplexity API
  help        Help about any command
  query       
  version     print version of pplx

Flags:
  -h, --help   help for pplx

Use "pplx [command] --help" for more information about a command.
```

## Chat

Chat with the Perplexity API.

```sh
pplx chat
```

## Query

Query the Perplexity API.

```sh
pplx query -p "what are the best citations of Jean Marc Jancovici ?" -s "you're a politician"
```

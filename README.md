# pplx

It's an unofficial CLI program to query/chat with the [perplexity API](https://www.perplexity.ai/).

## Installation

Download the latest release from the [releases page](https://github.com/sgaunet/pplx/releases).

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

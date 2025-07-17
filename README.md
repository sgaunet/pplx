# pplx

[![GitHub release](https://img.shields.io/github/release/sgaunet/pplx.svg)](https://github.com/sgaunet/pplx/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/sgaunet/pplx)](https://goreportcard.com/report/github.com/sgaunet/pplx)
![GitHub Downloads](https://img.shields.io/github/downloads/sgaunet/pplx/total)
[![Snapshot](https://github.com/sgaunet/pplx/actions/workflows/snapshot.yml/badge.svg)](https://github.com/sgaunet/pplx/actions/workflows/snapshot.yml)
[![Release](https://github.com/sgaunet/pplx/actions/workflows/release.yml/badge.svg)](https://github.com/sgaunet/pplx/actions/workflows/release.yml)
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
brew install sgaunet/tools/pplx
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

The above command will return in console a result that looks like:

![pplx query](img/cli.png)

### Query Examples

#### Basic Queries

```sh
# Simple query
pplx query -p "What is the capital of France?"

# Query with system prompt
pplx query -p "Explain quantum computing" -s "You are a physics professor"

# Query with custom model
pplx query -p "Latest AI news" --model "llama-3.1-sonar-large-128k-online"
```

#### Advanced Search Options

```sh
# Search only from specific domains
pplx query -p "climate change research" --search-domains nature.com,science.org

# Get recent information only (last week)
pplx query -p "stock market updates" --search-recency week

# Location-based query
pplx query -p "weather forecast" --location-lat 48.8566 --location-lon 2.3522 --location-country FR
```

#### Response Enhancement

```sh
# Include images in the response
pplx query -p "Famous landmarks in Paris" --return-images

# Get related questions
pplx query -p "How to learn programming" --return-related

# Filter images by format and domain
pplx query -p "Nature photography" --return-images --image-formats jpg,png --image-domains unsplash.com,pexels.com
```

#### Generation Parameters

```sh
# Control response length
pplx query -p "Summarize War and Peace" --max-tokens 500

# Fine-tune creativity and randomness
pplx query -p "Write a haiku about coding" --temperature 0.8 --top-p 0.95

# Adjust frequency and presence penalties
pplx query -p "Explain machine learning concepts" --frequency-penalty 0.5 --presence-penalty 0.3
```

#### Combined Examples

```sh
# Technical research with specific sources and recent data
pplx query -p "Latest developments in quantum computing" \
  --search-domains arxiv.org,nature.com \
  --search-recency month \
  --return-related \
  --max-tokens 1000

# Local business search with images
pplx query -p "Best restaurants near me" \
  --location-lat 40.7128 \
  --location-lon -74.0060 \
  --location-country US \
  --return-images \
  --search-recency week

# Creative writing with custom parameters
pplx query -p "Write a short story about AI" \
  -s "You are a creative science fiction writer" \
  --temperature 0.9 \
  --top-k 50 \
  --max-tokens 2000
```

## Available Options

### Common Options (for both chat and query)

| Option | Short | Type | Description |
|--------|-------|------|-------------|
| `--model` | `-m` | string | AI model to use |
| `--frequency-penalty` | | float64 | Penalize frequent tokens (0.0-2.0) |
| `--max-tokens` | | int | Maximum tokens in response |
| `--presence-penalty` | | float64 | Penalize already present tokens (0.0-2.0) |
| `--temperature` | | float64 | Response randomness (0.0-2.0) |
| `--top-k` | | int | Consider only top K tokens |
| `--top-p` | | float64 | Nucleus sampling threshold |
| `--timeout` | | duration | HTTP request timeout |
| `--search-domains` | | []string | Filter search to specific domains |
| `--search-recency` | | string | Filter by time: day, week, month, year |
| `--location-lat` | | float64 | User location latitude |
| `--location-lon` | | float64 | User location longitude |
| `--location-country` | | string | User location country code |
| `--return-images` | | bool | Include images in response |
| `--return-related` | | bool | Include related questions |
| `--stream` | | bool | Enable streaming responses |
| `--image-domains` | | []string | Filter images by domains |
| `--image-formats` | | []string | Filter images by formats |

### Query-specific Options

| Option | Short | Type | Description |
|--------|-------|------|-------------|
| `--user-prompt` | `-p` | string | User question/prompt (required) |
| `--sys-prompt` | `-s` | string | System prompt to set AI behavior |

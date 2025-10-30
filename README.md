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
# Search only from specific domains (using short flag)
pplx query -p "climate change research" -d nature.com,science.org

# Get recent information only (last week) - using short flag
pplx query -p "stock market updates" -r week

# Location-based query
pplx query -p "weather forecast" --location-lat 48.8566 --location-lon 2.3522 --location-country FR
```

#### Response Enhancement

```sh
# Include images in the response (using short flag)
pplx query -p "Famous landmarks in Paris" -i

# Get related questions (using short flag)
pplx query -p "How to learn programming" -q

# Filter images by format and domain
pplx query -p "Nature photography" -i --image-formats jpg,png --image-domains unsplash.com,pexels.com
```

#### Generation Parameters

```sh
# Control response length (using short flag)
pplx query -p "Summarize War and Peace" -T 500

# Fine-tune creativity and randomness (using short flags)
pplx query -p "Write a haiku about coding" -t 0.8 --top-p 0.95

# Adjust frequency and presence penalties
pplx query -p "Explain machine learning concepts" --frequency-penalty 0.5 --presence-penalty 0.3
```

#### Combined Examples

```sh
# Technical research with specific sources and recent data (using short flags)
pplx query -p "Latest developments in quantum computing" \
  -d arxiv.org,nature.com \
  -r month \
  -q \
  -T 1000

# Local business search with images (using short flags)
pplx query -p "Best restaurants near me" \
  --location-lat 40.7128 \
  --location-lon -74.0060 \
  --location-country US \
  -i \
  -r week

# Creative writing with custom parameters (using short flags)
pplx query -p "Write a short story about AI" \
  -s "You are a creative science fiction writer" \
  -t 0.9 \
  -k 50 \
  -T 2000
```

## Available Options

### Common Options (for both chat and query)

| Option | Short | Type | Description |
|--------|-------|------|-------------|
| `--model` | `-m` | string | AI model to use |
| `--frequency-penalty` | | float64 | Penalize frequent tokens (0.0-2.0) |
| `--max-tokens` | `-T` | int | Maximum tokens in response |
| `--presence-penalty` | | float64 | Penalize already present tokens (0.0-2.0) |
| `--temperature` | `-t` | float64 | Response randomness (0.0-2.0) |
| `--top-k` | `-k` | int | Consider only top K tokens |
| `--top-p` | | float64 | Nucleus sampling threshold |
| `--timeout` | | duration | HTTP request timeout |
| `--search-domains` | `-d` | []string | Filter search to specific domains |
| `--search-recency` | `-r` | string | Filter by time: day, week, month, year |
| `--search-mode` | `-a` | string | Search mode: web (default) or academic |
| `--search-context-size` | `-c` | string | Search context size: low, medium, or high |
| `--location-lat` | | float64 | User location latitude |
| `--location-lon` | | float64 | User location longitude |
| `--location-country` | | string | User location country code |
| `--return-images` | `-i` | bool | Include images in response |
| `--return-related` | `-q` | bool | Include related questions |
| `--stream` | `-S` | bool | Enable streaming responses |
| `--image-domains` | | []string | Filter images by domains |
| `--image-formats` | | []string | Filter images by formats |

### Query-specific Options

| Option | Short | Type | Description |
|--------|-------|------|-------------|
| `--user-prompt` | `-p` | string | User question/prompt (required) |
| `--sys-prompt` | `-s` | string | System prompt to set AI behavior |

## MCP Server (Model Context Protocol)

The `pplx mcp-stdio` command provides an MCP server that exposes Perplexity AI functionality to Claude Code and other MCP-compatible clients.

### Quick Start with Claude Code

```bash
# Install the server
brew tap sgaunet/homebrew-tools
brew install sgaunet/tools/pplx
# or download from releases and place in PATH

# Add to Claude Code
claude mcp add perplexity-ai -s user -- pplx mcp-stdio
```

### Manual Configuration

#### For Claude Code (via config file)

Add to your Claude Code MCP configuration:

```json
{
  "mcpServers": {
    "perplexity-ai": {
      "command": "pplx",
      "args": ["mcp-stdio"],
      "env": {
        "PPLX_API_KEY": "your_perplexity_api_key_here"
      }
    }
  }
}
```

Configure `PPLX_API_KEY` with your Perplexity AI API key in your environment or directly in the MCP config.

#### For Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or equivalent location:

```json
{
  "mcpServers": {
    "perplexity-ai": {
      "command": "/usr/local/bin/pplx",
      "args": ["mcp-stdio"],
      "env": {
        "PPLX_API_KEY": "your_perplexity_api_key_here"
      }
    }
  }
}
```

#### For Other MCP Clients

Any MCP-compatible client can use this server by executing:

```bash
PPLX_API_KEY=your_key /path/to/pplx mcp-stdio
```

### MCP Tool: `query`

The MCP server exposes a single powerful tool called `query` with the following parameters:

#### Required Parameters
- `user_prompt` (string): The user question/prompt

#### Optional Parameters

**Core Parameters:**
- `system_prompt` (string): System prompt to guide AI behavior
- `model` (string): AI model to use (default: sonar-small-online)
- `temperature` (number): Response randomness (0.0-2.0)
- `max_tokens` (number): Maximum tokens in response
- `frequency_penalty` (number): Penalize frequent tokens (0.0-2.0)
- `presence_penalty` (number): Penalize already present tokens (0.0-2.0)
- `top_k` (number): Consider only top K tokens
- `top_p` (number): Nucleus sampling threshold
- `timeout` (number): HTTP timeout in seconds

**Search & Web Options:**
- `search_domains` (array): Filter search to specific domains
- `search_recency` (string): Filter by time: "day", "week", "month", "year", "hour"
- `location_lat` (number): User location latitude
- `location_lon` (number): User location longitude
- `location_country` (string): User location country code
- `search_mode` (string): Search mode: "web" or "academic"
- `search_context_size` (string): Context size: "low", "medium", "high"

**Response Enhancement:**
- `return_images` (boolean): Include images in response
- `return_related` (boolean): Include related questions
- `stream` (boolean): Enable streaming (collected into complete response)

**Image Filtering:**
- `image_domains` (array): Filter images by domains
- `image_formats` (array): Filter images by formats (jpg, png, etc.)

**Response Formats (Sonar models only):**
- `response_format_json_schema` (string): JSON schema for structured output
- `response_format_regex` (string): Regex pattern for structured output

**Date Filtering:**
- `search_after_date` (string): Filter results published after date (MM/DD/YYYY)
- `search_before_date` (string): Filter results published before date (MM/DD/YYYY)
- `last_updated_after` (string): Filter results last updated after date (MM/DD/YYYY)
- `last_updated_before` (string): Filter results last updated before date (MM/DD/YYYY)

**Deep Research:**
- `reasoning_effort` (string): For sonar-deep-research model: "low", "medium", "high"

### Example Usage in Claude Code

Once configured, you can use the Perplexity MCP server directly in Claude Code:

```
Search for recent AI developments in computer vision with images
```

Claude Code will automatically use the MCP server to:
1. Query Perplexity AI with your prompt
2. Filter for recent information
3. Include relevant images
4. Return structured results with citations

### Response Format

The MCP server returns JSON with the following structure:

```json
{
  "content": "AI response text with markdown formatting",
  "model": "model_used_for_generation",
  "usage": {
    "prompt_tokens": 123,
    "completion_tokens": 456,
    "total_tokens": 579
  },
  "search_results": [
    {
      "title": "Article Title",
      "url": "https://example.com/article",
      "snippet": "Relevant excerpt..."
    }
  ],
  "images": [
    {
      "url": "https://example.com/image.jpg",
      "description": "Image description"
    }
  ],
  "related_questions": [
    "What are the latest AI breakthroughs?",
    "How is computer vision evolving?"
  ]
}
```

### Environment Variables

- `PPLX_API_KEY` (required): Your Perplexity AI API key

### Troubleshooting

1. **Server not starting**: Verify `PPLX_API_KEY` is set
2. **Command not found**: Ensure `pplx` is in your PATH
3. **Configuration issues**: Check JSON syntax in MCP config files
4. **API errors**: Verify your API key is valid and has sufficient credits

### Advanced Configuration Examples

#### High-quality research with academic sources
```json
{
  "user_prompt": "Latest quantum computing breakthroughs",
  "search_mode": "academic",
  "search_context_size": "high",
  "search_recency": "month",
  "return_related": true,
  "max_tokens": 2000
}
```

#### Location-based search with images
```json
{
  "user_prompt": "Best restaurants in Tokyo",
  "location_lat": 35.6762,
  "location_lon": 139.6503,
  "location_country": "JP",
  "return_images": true,
  "image_formats": ["jpg", "png"],
  "search_recency": "week"
}
```

#### Structured output for data processing
```json
{
  "user_prompt": "List top 5 programming languages",
  "model": "sonar-small-online",
  "response_format_json_schema": "{\"type\":\"object\",\"properties\":{\"languages\":{\"type\":\"array\",\"items\":{\"type\":\"string\"}}}}",
  "max_tokens": 500
}
```

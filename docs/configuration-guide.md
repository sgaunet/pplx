# Configuration Guide

Complete reference for configuring the `pplx` CLI tool.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Configuration Options Reference](#configuration-options-reference)
3. [Templates](#templates)
4. [Profiles](#profiles)
5. [Commands](#commands)
6. [Environment Variables](#environment-variables)
7. [Precedence Rules](#precedence-rules)

## Quick Start

### Interactive Setup (Recommended)

The easiest way to get started is with the interactive configuration wizard:

```bash
pplx config init --interactive
```

This guides you through all configuration options with helpful prompts.

### Template-Based Setup

Start quickly with pre-configured templates:

```bash
# For academic research
pplx config init --template research

# For creative writing
pplx config init --template creative

# For news and current events
pplx config init --template news
```

### Manual Setup

Create a minimal configuration:

```bash
pplx config init
```

Configuration is stored at `~/.config/pplx/config.yaml`.

## Configuration Options Reference

The configuration file has four main sections: `defaults`, `search`, `output`, and `api`.

### Defaults Section

Controls model behavior and response generation parameters.

| Option | Type | Default | Valid Range | Description |
|--------|------|---------|-------------|-------------|
| `model` | string | `"sonar"` | Any Perplexity model | Model to use for queries. Options: `sonar`, `sonar-pro`, `sonar-deep-research` |
| `temperature` | float | `0.2` | `0.0` - `1.0` | Controls randomness. Lower = more deterministic, Higher = more creative |
| `max_tokens` | int | `4000` | `1` - `16384` | Maximum number of tokens in the response (model-dependent) |
| `top_k` | int | `0` | `0` - `100` | Top-K sampling: limit to K highest probability tokens. `0` disables |
| `top_p` | float | `0.0` | `0.0` - `1.0` | Top-P (nucleus) sampling: cumulative probability threshold |
| `frequency_penalty` | float | `0.0` | `0.0` - `2.0` | Reduces repetition of frequent tokens |
| `presence_penalty` | float | `0.0` | `0.0` - `2.0` | Encourages discussing new topics |
| `timeout` | string | `"120s"` | Duration string | Timeout for API requests (e.g., `"30s"`, `"5m"`) |

**Example:**
```yaml
defaults:
  model: sonar
  temperature: 0.7
  max_tokens: 4096
  top_p: 0.9
  timeout: 120s
```

### Search Section

Controls search behavior and filtering.

| Option | Type | Default | Valid Values | Description |
|--------|------|---------|--------------|-------------|
| `domains` | []string | `[]` | Domain list | Limit search to specific domains (e.g., `["wikipedia.org", "github.com"]`) |
| `recency` | string | `""` | `hour`, `day`, `week`, `month`, `year` | Time-based filtering of search results |
| `mode` | string | `"web"` | `web`, `academic` | Search mode: general web or academic sources |
| `context_size` | string | `"medium"` | `low`, `medium`, `high` | Amount of context from search results |
| `location_lat` | float | `0.0` | `-90.0` to `90.0` | Geographic location latitude |
| `location_lon` | float | `0.0` | `-180.0` to `180.0` | Geographic location longitude |
| `location_country` | string | `""` | ISO 3166-1 alpha-2 | Country code (e.g., `"US"`, `"GB"`) |
| `after_date` | string | `""` | `MM/DD/YYYY` | Filter results published after this date |
| `before_date` | string | `""` | `MM/DD/YYYY` | Filter results published before this date |
| `last_updated_after` | string | `""` | `MM/DD/YYYY` | Filter results last updated after this date |
| `last_updated_before` | string | `""` | `MM/DD/YYYY` | Filter results last updated before this date |

**Example:**
```yaml
search:
  mode: academic
  context_size: high
  recency: week
  domains:
    - scholar.google.com
    - arxiv.org
    - pubmed.ncbi.nlm.nih.gov
```

### Output Section

Controls response format and content.

| Option | Type | Default | Valid Values | Description |
|--------|------|---------|--------------|-------------|
| `stream` | bool | `false` | `true`, `false` | Enable streaming responses (output tokens as generated) |
| `return_images` | bool | `false` | `true`, `false` | Include images in the response |
| `return_related` | bool | `false` | `true`, `false` | Include related questions in the response |
| `json` | bool | `false` | `true`, `false` | Output response as JSON instead of formatted text |
| `image_domains` | []string | `[]` | Domain list | Filter images by domain (e.g., `["unsplash.com"]`) |
| `image_formats` | []string | `[]` | Format list | Filter images by format: `jpg`, `png`, `gif`, `webp` |
| `response_format_json_schema` | string | `""` | JSON schema | JSON schema for structured output (sonar model only) |
| `response_format_regex` | string | `""` | Regex pattern | Regex pattern for structured output (sonar model only) |
| `reasoning_effort` | string | `""` | `low`, `medium`, `high` | Reasoning effort for sonar-deep-research model |

**Example:**
```yaml
output:
  stream: true
  return_images: true
  return_related: true
  json: false
```

### API Section

API connection settings.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `key` | string | `""` | API key (typically set via `PERPLEXITY_API_KEY` environment variable) |
| `base_url` | string | `""` | Custom API base URL (if using a proxy or custom endpoint) |
| `timeout` | duration | `0s` | API request timeout (Go duration format) |

**Example:**
```yaml
api:
  key: ${PERPLEXITY_API_KEY}
  timeout: 30s
```

## Templates

Templates provide pre-configured settings for common use cases. Choose a template with `pplx config init --template <name>`.

### Template Comparison

| Feature | Research | Creative | News | Full-Example |
|---------|----------|----------|------|--------------|
| **Model** | sonar | sonar | sonar | sonar |
| **Temperature** | 0.3 (factual) | 0.9 (creative) | 0.5 (balanced) | 0.7 (default) |
| **Max Tokens** | 4096 | 8192 | 4096 | 4096 |
| **Search Mode** | academic | web | web | web |
| **Context Size** | high | medium | medium | medium |
| **Streaming** | No | Yes | No | No |
| **Return Images** | No | Yes | Yes | No |
| **Return Related** | Yes | Yes | Yes | No |
| **Domain Focus** | Academic sources | General web | News outlets | None |
| **Recency Filter** | None | None | week | None |
| **Best For** | Academic research, scholarly sources | Creative writing, brainstorming | Current events, news | Learning all options |

### Research Template

Optimized for academic and scholarly research with authoritative sources.

**Key Settings:**
- Lower temperature (0.3) for factual, consistent responses
- Academic search mode
- High context size for comprehensive research
- Focus on scholarly domains (scholar.google.com, arxiv.org, pubmed, etc.)

**Use Cases:**
- Literature reviews
- Academic paper research
- Technical documentation research
- Scientific queries

### Creative Template

Optimized for creative writing, brainstorming, and exploratory queries.

**Key Settings:**
- High temperature (0.9) for creative, diverse responses
- Larger token limit (8192) for longer creative output
- Frequency and presence penalties to encourage diversity
- Streaming enabled for real-time creative output
- Images included for visual inspiration

**Use Cases:**
- Creative writing
- Brainstorming sessions
- Content ideation
- Exploratory research

### News Template

Optimized for current events and recent news coverage.

**Key Settings:**
- Balanced temperature (0.5)
- Week recency filter for recent news
- Focus on reputable news outlets (Reuters, BBC, NYT, etc.)
- Images included for news photos

**Use Cases:**
- Breaking news research
- Current events analysis
- News aggregation
- Topical research

### Full-Example Template

Comprehensive template showing all available options with detailed comments and descriptions. Use this as a reference or starting point for custom configurations.

## Profiles

Profiles allow you to maintain multiple named configurations and switch between them easily.

### Creating Profiles

Profiles can be created in your config file or via commands:

```yaml
profiles:
  research:
    name: research
    description: Academic research settings
    defaults:
      model: sonar
      temperature: 0.3
      max_tokens: 4096
    search:
      mode: academic
      context_size: high
    output:
      return_related: true

  creative:
    name: creative
    description: Creative writing settings
    defaults:
      model: sonar
      temperature: 0.9
      max_tokens: 8192
    output:
      stream: true
      return_images: true

active_profile: research
```

### Profile Management Commands

```bash
# List all profiles
pplx config profile list

# Create a new profile
pplx config profile create my-profile

# Switch to a profile
pplx config profile switch research

# Delete a profile
pplx config profile delete old-profile

# Show specific profile
pplx config show --profile research
```

### Using Profiles

When a profile is active, its settings override the default configuration. You can still override profile settings with CLI flags:

```bash
# Use creative profile but with different temperature
pplx --temperature 0.5 "write a story"
```

## Commands

### config init

Initialize a new configuration file.

**Usage:**
```bash
pplx config init [flags]
```

**Flags:**
- `-t, --template <name>`: Use a template (research, creative, news, full-example)
- `-i, --interactive`: Launch interactive configuration wizard
- `--with-examples`: Include example configurations and annotations
- `--with-profiles`: Include profile configurations
- `-f, --force`: Force overwrite existing configuration
- `--check-env`: Check environment variables (API keys)

**Examples:**

```bash
# Interactive wizard (recommended for first-time setup)
pplx config init --interactive

# Use research template
pplx config init --template research

# Create annotated config with examples
pplx config init --with-examples

# Force overwrite existing config with news template
pplx config init --template news --force

# Check environment and create config
pplx config init --check-env

# Combine options
pplx config init --template creative --with-examples --check-env
```

### config show

Display the current configuration.

**Usage:**
```bash
pplx config show [flags]
```

**Flags:**
- `--json`: Output in JSON format
- `--profile <name>`: Show specific profile

**Examples:**

```bash
# Show current configuration
pplx config show

# Show in JSON format
pplx config show --json

# Show specific profile
pplx config show --profile research
```

### config validate

Validate configuration file syntax and values.

**Usage:**
```bash
pplx config validate
```

**Examples:**

```bash
# Validate current configuration
pplx config validate

# Validate specific file
pplx config validate --config /path/to/config.yaml
```

### config edit

Open configuration file in your default editor.

**Usage:**
```bash
pplx config edit
```

The command uses the `$EDITOR` environment variable (defaults to `vi`). After editing, the configuration is automatically validated.

**Examples:**

```bash
# Edit configuration
pplx config edit

# Edit specific file
pplx config edit --config /path/to/config.yaml
```

### config path

Show configuration file location and search paths.

**Usage:**
```bash
pplx config path [flags]
```

**Flags:**
- `-c, --check`: Validate configuration and show details

**Examples:**

```bash
# Show config file location
pplx config path

# Show location and validate
pplx config path --check
```

**Output:**
```
Configuration File Search Order:

📁 ~/.config/pplx/
   ✓  config.yaml
   ⚪ pplx.yaml
   ⚪ config.yml
   ⚪ pplx.yml

✓ Active configuration: ~/.config/pplx/config.yaml
```

### config options

List all available configuration options with metadata.

**Usage:**
```bash
pplx config options [flags]
```

**Flags:**
- `-s, --section <name>`: Filter by section (defaults, search, output, api)
- `-f, --format <format>`: Output format (table, json, yaml)
- `-v, --validation`: Show validation rules

**Examples:**

```bash
# List all options in table format
pplx config options

# List in JSON format
pplx config options --format json

# List only search options
pplx config options --section search

# Show with validation rules
pplx config options --validation

# Combine filters
pplx config options --section defaults --format yaml --validation
```

### config profile

Manage configuration profiles.

**Subcommands:**
- `list`: List all profiles
- `create <name>`: Create a new profile
- `switch <name>`: Switch to a profile
- `delete <name>`: Delete a profile

**Examples:**

```bash
# List all profiles
pplx config profile list

# Create new profile
pplx config profile create development

# Switch to profile
pplx config profile switch research

# Delete profile
pplx config profile delete old-profile
```

## Environment Variables

Configuration supports environment variable interpolation using `${VAR_NAME}` syntax.

### Available Environment Variables

| Variable | Description |
|----------|-------------|
| `PERPLEXITY_API_KEY` | Your Perplexity API key (required) |
| `PERPLEXITY_BASE_URL` | Custom API base URL (optional) |
| `EDITOR` | Default editor for `config edit` command |

### Using Environment Variables in Config

```yaml
api:
  key: ${PERPLEXITY_API_KEY}
  base_url: ${PERPLEXITY_BASE_URL}

defaults:
  model: ${PPLX_MODEL:-sonar}  # Default to "sonar" if not set
```

### Setting Environment Variables

**Bash/Zsh:**
```bash
export PERPLEXITY_API_KEY="your-api-key-here"
```

**Fish:**
```fish
set -x PERPLEXITY_API_KEY "your-api-key-here"
```

**Persistent (add to shell config):**
```bash
# ~/.bashrc or ~/.zshrc
export PERPLEXITY_API_KEY="your-api-key-here"
```

## Precedence Rules

Configuration values are determined by the following order (highest to lowest priority):

1. **CLI Flags** - Explicit command-line flags always take precedence
2. **Environment Variables** - Environment variables override config file
3. **Active Profile** - Active profile settings override defaults
4. **Configuration File** - Values from config file
5. **Built-in Defaults** - Hardcoded defaults if nothing else specified

### Example

Given this configuration:
```yaml
defaults:
  model: sonar
  temperature: 0.3

profiles:
  creative:
    defaults:
      temperature: 0.9

active_profile: creative
```

And environment variable:
```bash
export PPLX_TEMPERATURE=0.5
```

Running:
```bash
pplx --temperature 0.7 "query"
```

**Effective temperature:** `0.7` (from CLI flag)

If run without the `--temperature` flag:
```bash
pplx "query"
```

**Effective temperature:** `0.5` (from environment variable)

If no environment variable and no CLI flag:
**Effective temperature:** `0.9` (from active profile)

## Configuration Discovery

The `pplx` tool searches for configuration files in the following locations (in order):

1. `~/.config/pplx/config.yaml`
2. `~/.config/pplx/pplx.yaml`
3. `~/.config/pplx/config.yml`
4. `~/.config/pplx/pplx.yml`

The first file found is used. Use `pplx config path` to see which file is active.

## Validation

Configuration validation ensures:

- Temperature is between 0.0 and 1.0
- Max tokens is positive and within model limits
- Top-K is between 0 and 100
- Top-P is between 0.0 and 1.0
- Penalties are between 0.0 and 2.0
- Dates are in MM/DD/YYYY format
- Location coordinates are valid
- Search mode is either "web" or "academic"
- Context size is "low", "medium", or "high"
- Recency is one of: hour, day, week, month, year
- Image formats are valid: jpg, png, gif, webp
- Reasoning effort is "low", "medium", or "high"

Run `pplx config validate` to check your configuration.

## Troubleshooting

### Configuration Not Found

```bash
# Check search paths and create config
pplx config path
pplx config init
```

### Invalid Configuration

```bash
# Validate and see errors
pplx config validate

# Edit and fix
pplx config edit
```

### API Key Issues

```bash
# Check environment
pplx config init --check-env

# Set API key
export PERPLEXITY_API_KEY="your-key"

# Or add to config
pplx config edit
# Then add:
# api:
#   key: ${PERPLEXITY_API_KEY}
```

### Profile Not Working

```bash
# List profiles
pplx config profile list

# Switch to correct profile
pplx config profile switch profile-name

# Verify active profile
pplx config show | grep active_profile
```

## Advanced Usage

### Multiple Configuration Files

Use the `--config` flag to specify different configuration files:

```bash
pplx --config ~/.config/pplx/research.yaml "query"
pplx --config ~/.config/pplx/creative.yaml "query"
```

### Programmatic Configuration

For integration with other tools, output configuration as JSON:

```bash
pplx config show --json > config.json
```

### Template Customization

Start with a template and customize:

```bash
# Create from template
pplx config init --template research

# Edit to customize
pplx config edit
```

### Shell Completion

Enable shell completion for config commands:

```bash
# Bash
pplx completion bash > /etc/bash_completion.d/pplx

# Zsh
pplx completion zsh > ~/.zsh/completion/_pplx

# Fish
pplx completion fish > ~/.config/fish/completions/pplx.fish
```

## Related Documentation

- [README.md](../README.md) - Quick start and basic usage
- [API Documentation](https://docs.perplexity.ai) - Perplexity API reference
- [Template Files](../pkg/config/templates/) - Template source files

## Need Help?

- Check configuration: `pplx config show`
- Validate configuration: `pplx config validate`
- List all options: `pplx config options`
- View this guide: `docs/configuration-guide.md`

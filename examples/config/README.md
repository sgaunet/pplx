# Configuration Examples

This directory contains example configuration files for `pplx`. These examples demonstrate various use cases and configuration patterns.

## Available Examples

### Quick Start Templates

- **`full-example.yaml`** - Comprehensive reference showing all 29+ available options with detailed comments
  - Best for: Learning what's possible, creating custom configurations
  - Use with: `pplx config init --template full-example`

### Use Case Templates

- **`research.yaml`** - Academic and scholarly research configuration
  - Optimized for: Research papers, academic sources, authoritative content
  - Features: Academic search mode, high context, verified domains
  - Use with: `pplx config init --template research`

- **`creative.yaml`** - Creative writing and brainstorming configuration
  - Optimized for: Content generation, creative tasks, exploratory queries
  - Features: High temperature, streaming output, diverse sources
  - Use with: `pplx config init --template creative`

- **`news.yaml`** - News and current events configuration
  - Optimized for: Latest news, current events, trending topics
  - Features: Reputable news sources, weekly recency filter
  - Use with: `pplx config init --template news`

## Using Templates

### Option 1: Initialize with Template

Create a new config file using a template:

```bash
# Create config from a template
pplx config init --template research

# Create with examples and env check
pplx config init --template creative --with-examples --check-env

# Interactive wizard
pplx config init --interactive
```

### Option 2: Copy Manually

Copy any example file to your config location:

```bash
# Copy to default config location
cp examples/config/research.yaml ~/.config/pplx/config.yaml

# Or to a custom location
cp examples/config/creative.yaml ./my-config.yaml
pplx --config ./my-config.yaml query "your question"
```

## Configuration File Locations

Config files are loaded from these locations (in order of precedence):

1. Path specified with `--config` flag
2. `PPLX_CONFIG_FILE` environment variable
3. `./pplx.yaml` (current directory)
4. `~/.config/pplx/config.yaml` (user config directory)
5. `/etc/pplx/config.yaml` (system-wide config)

## Security Note

⚠️ **Protect your API keys!**

- Never commit config files with API keys to version control
- Use environment variables: `export PERPLEXITY_API_KEY=your-key`
- Set proper file permissions: `chmod 600 ~/.config/pplx/config.yaml`
- The tool will warn you if permissions are too permissive

## Customization

All templates can be customized for your needs:

1. Start with a template that matches your use case
2. Adjust parameters like temperature, max_tokens, etc.
3. Add/remove search domains
4. Configure output preferences
5. Save and test with `pplx query "test question"`

## Available Options

See `full-example.yaml` for complete documentation of all options, including:

- **Model settings**: model, temperature, max_tokens, top_p, top_k
- **Search filters**: domains, recency, mode, context_size, location
- **Date filtering**: after_date, before_date, last_updated filters
- **Output format**: streaming, images, related questions, JSON
- **Response control**: JSON schema, regex patterns, reasoning effort

## Template Maintenance

These templates are synchronized with the embedded templates in the `pplx` binary. They are automatically updated when new options are added to ensure examples remain current.

To verify templates are up to date:

```bash
# List available templates
pplx config init --help

# Generate a new config from latest template
pplx config init --template full-example --force
```

## Learn More

- Run `pplx config --help` for all configuration commands
- Run `pplx config show` to view your active configuration
- Run `pplx config init --check-env` to verify API key setup
- Visit the [Perplexity API Documentation](https://docs.perplexity.ai/) for model details

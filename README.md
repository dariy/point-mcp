# Point MCP Server

`point-mcp` is a [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server that provides tools, resources, and prompts for interacting with the Point CMS API. It allows AI models to manage posts, media, tags, and themes, as well as retrieve blog context and analytics.

## Features

- **Post Management**: Create, update, delete, list, publish, and withdraw posts.
- **Media Library**: Upload local files, list media items, and perform AI-driven analysis of images.
- **Taxonomy & Design**: Manage tags and switch between blog themes.
- **Blog Context**: Retrieve site-wide settings, active theme metadata, and content statistics.
- **Syntax Guidelines**: Access detailed rules for Point CMS markdown extensions, allowed HTML, and CSS restrictions.
- **Analytics**: Get top-performing posts and aggregate view statistics.
- **Landing Page Workflow**: A guided prompt for creating theme-aware immersive landing pages.

## Configuration

The server is configured via environment variables, which can also be provided in a `.env` file in the working directory.

| Variable | Description | Default |
|----------|-------------|---------|
| `POINT_BASE_URL` | **Required**. The base URL of your Point CMS API. | - |
| `POINT_API_KEY` | **Required**. Your Point API key. | - |
| `POINT_MCP_TRANSPORT` | Transport method for MCP: `stdio` or `http`. | `stdio` |
| `POINT_MCP_HTTP_ADDR` | Listen address when using `http` transport. | `:9000` |
| `POINT_MCP_LOG_FILE` | Optional path to a file where logs should be written. | - |
| `MCP_AUTH_TOKENS` | **Required in HTTP mode**. Comma-separated bearer tokens for client authentication. Supports multiple tokens for zero-downtime rotation. | - |

## Installation

### Prerequisites

- [Go](https://go.dev/) 1.26 or later.

### Build

Run the provided build script to compile the binary:

```bash
./build.sh
```

This will create a `point-mcp` binary in the root directory.

### Docker

Pre-built images are published to the GitHub Container Registry on every push to `main`:

```bash
docker pull ghcr.io/dariy/point-mcp:main
```

Run with Docker:

```bash
docker run -d --restart unless-stopped \
  -p 127.0.0.1:9000:9000 \
  -e POINT_MCP_TRANSPORT=http \
  -e POINT_MCP_HTTP_ADDR=0.0.0.0:9000 \
  -e MCP_AUTH_TOKENS=your-secret-token \
  -e POINT_BASE_URL=https://your-point-instance.com \
  -e POINT_API_KEY=your-api-key \
  ghcr.io/dariy/point-mcp:main
```

## Usage

### Running as an MCP Server (stdio)

This is the default mode, typically used when connecting the server to an MCP-compatible IDE or client (like Claude Desktop).

```bash
./point-mcp
```

### Running as an MCP Server (HTTP)

To start the server using the streamable-HTTP transport, `MCP_AUTH_TOKENS` must be set. Generate a token with `openssl rand -hex 32`.

```bash
POINT_MCP_TRANSPORT=http MCP_AUTH_TOKENS=your-secret-token ./point-mcp
```

All HTTP requests must include an `Authorization: Bearer <token>` header. The server rejects unauthenticated requests with `401 Unauthorized`.

**Token rotation** (zero-downtime): set multiple comma-separated tokens, update clients one by one, then remove the old token.

```bash
MCP_AUTH_TOKENS=old-token,new-token  # both accepted during rotation
```

### Connecting Claude or Gemini

Configure your MCP client to use the server's URL with a bearer token header.

**Claude Code** (`~/.claude/settings.json`):

```json
{
  "mcpServers": {
    "point-mcp": {
      "type": "sse",
      "url": "https://mcp.example.com/",
      "headers": { "Authorization": "Bearer your-secret-token" }
    }
  }
}
```

**Gemini**:

```json
{
  "servers": [{
    "name": "point-mcp",
    "url": "https://mcp.example.com/",
    "headers": { "Authorization": "Bearer your-secret-token" }
  }]
}
```

### Running as a CLI Tool

You can call any tool directly from the command line for testing or automation:

```bash
./point-mcp point_get_context
./point-mcp point_list_posts '{"status": "published", "per_page": 5}'
```

## MCP Elements

### Tools

- **Context & Guidelines**:
    - `point_get_context`: Returns site title, author, stats, and active theme info.
    - `point_get_theme_css`: Returns active theme CSS variables for harmonization.
    - `point_get_syntax_guidelines`: Returns rules for markdown, HTML, and CSS.
- **Posts**:
    - `point_list_posts`: List posts with filters (status, tag, search, sort).
    - `point_get_post`: Fetch a single post by ID or slug.
    - `point_create_post`: Create a new draft or published post.
    - `point_update_post`: Update an existing post.
    - `point_publish_post` / `point_withdraw_post`: Change publication status.
    - `point_delete_post`: Permanently remove a post.
    - `point_generate_preview_link`: Get a temporary URL for a draft.
    - `point_replace_in_post`: Efficiently replace text within a specific field.
- **Media**:
    - `point_upload_media`: Upload a local file to the library.
    - `point_list_media`: Browse the media library.
    - `point_analyze_media`: Get AI-suggested title and tags for an image.
- **Tags & Themes**:
    - `point_list_tags` / `point_create_tag`: Manage blog tags.
    - `point_list_themes` / `point_set_active_theme`: Browse and switch themes.
- **Analytics**:
    - `point_analytics_top_posts`: Get most-viewed posts.
    - `point_analytics_summary`: Get site-wide view statistics.

### Resources

- `point://context`: Site-wide context and statistics (JSON).
- `point://theme/active`: Active theme CSS variables and `:root` block (JSON).
- `point://posts/recent`: A list of the most recently published posts (JSON).

### Prompts

- `create_landing_page`: A multi-step workflow that guides the model through discovering context, uploading media, authoring a theme-aware landing page, and generating a preview link.

## Development

- **Build Script**: `build.sh` - Compiles the `point-mcp` binary.
- **Inspection**: `inspect.sh` - Launches the [MCP Inspector](https://github.com/modelcontextprotocol/inspector) for interactive testing and debugging of the server.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

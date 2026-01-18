# yalms (Yet Another Logseq MCP Server)

`yalms` is a Model Context Protocol (MCP) server that provides a structured interface for interacting with Logseq graphs. It allows AI models to read, create, update, and manage pages and blocks within Logseq, supporting both general outliner usage and ontological data modeling.

## Features

- **Model Context Protocol (MCP)**: Implements the MCP standard for seamless integration with AI tools and IDEs.
- **Two Operation Modes**:
  - **General Mode**: Standard outliner behavior for managing pages and blocks.
  - **Ontological Mode**: Optimizes for structured data with namespacing support and snake_case property normalization.
- **Smart Link Management**: Automatically handles page creation for links and converts namespaced links to UUID references to maintain graph integrity.
- **Comprehensive API**: Supports searching, batch operations, tag management, and complex block tree manipulations.

## Installation

### Prerequisites

- Go 1.21 or later.
- A running Logseq instance with the [HTTP API](https://docs.logseq.com/#/page/HTTP%20API) enabled.

### Build from Source

```bash
git clone https://github.com/clstb/yalms.git
cd yalms
go build -o yalms ./cmd/yalms
```

## Usage

Start the server using the CLI:

```bash
./yalms [flags]
```

### Configuration Flags

| Flag | Environment Variable | Default | Description |
|------|----------------------|---------|-------------|
| `--logseq-url` | `LOGSEQ_URL` | `http://127.0.0.1:12315` | URL of the Logseq HTTP API. |
| `--logseq-token` | `LOGSEQ_TOKEN` | `auth` | API token for authentication. |
| `--logseq-mode` | `LOGSEQ_MODE` | `general` | Server mode: `general` or `ontological`. |
| `--debug` | - | `false` | Enable verbose development logging. |

## Available Tools

The server exposes several MCP tools depending on the active mode:

### Graph Tools
- `read_graph_info`: Get metadata about the current Logseq graph.
- `query`: Execute advanced Datalog queries against the Logseq database.
- `list_namespaces`: List all existing namespaces in the graph.
- `get_daily_journal`: Retrieve the page details for today's journal.

### Page/Entity Tools
- `read_page` (General) / `read_entity` (Ontological): Retrieve structured data and properties.
- `create_entity`: Create a new namespaced entity (Ontological) or page (General).
- `create_pages` (General): Create multiple pages in a single call.
- `update_page` (General) / `update_entity` (Ontological): Modify properties.
- `delete_page` (General) / `delete_entity` (Ontological): Permanently remove a page/entity.
- `delete_pages` (General): Permanently remove multiple pages.
- `rename_page`: Rename an existing page/entity by UUID.

### Namespace Tools
- `read_namespace`: List all entities or pages within a specific namespace.
- `create_namespace` (General): Create a new namespace/category level.

### Block/Entry Tools
- `read_block` (General) / `read_entry` (Ontological): Retrieve details for a specific block/entry.
- `create_block` (General) / `create_entry` (Ontological): Insert a single block/entry under a parent.
- `create_block_tree` (General) / `create_entry_tree` (Ontological): Insert a structured hierarchy.
- `append_block` (General) / `append_entry_to_entity` (Ontological): Add to the end of a page/entity.
- `update_block` (General) / `update_entry` (Ontological): Modify content or properties.
- `remove_block` (General) / `remove_entry` (Ontological): Remove a block/entry.
- `remove_blocks` (General): Remove multiple blocks.

### Tag/Property Tools
- `add_tag`: Add a `#tag` to a block/entry or page/entity (Class/Universal).
- `remove_tag`: Remove a discovery tag (Class/Universal).
- `add_property`: Add or update a specific metadata property (Attribute/Relationship).
- `remove_property`: Remove a specific metadata property (Attribute/Relationship).

## Ontological Mapping

`yalms` supports an ontological data model mapped to Logseq's native structures:

| Ontological Concept | Definition | Logseq Mapping | Example |
|---------------------|------------|----------------|---------|
| **Class (Universal)** | A category or type of thing. | Tag (linked page) | `#Book`, `#Person`, `#Meeting` |
| **Instance (Particular)** | A specific unique entity. | Page | `[[The Hobbit]]`, `[[Alice Smith]]` |
| **Attribute** | Data inherent to the instance. | Property (Text/Number) | `published-date:: 1937` |
| **Relationship** | Connection to other instances. | Property (Page Link) | `author:: [[J.R.R. Tolkien]]` |

## Development

### Running Tests

To run the standard unit tests:

```bash
go test ./...
```

To run integration tests (requires a running Logseq instance):

```bash
LOGSEQ_API_TOKEN=your_token go test ./pkg/logseq -v
```

### Project Structure

- `cmd/yalms`: CLI entry point.
- `internal/server`: MCP server implementation and request handlers.
- `pkg/logseq`: Logseq API client and data models.

## License

[MIT License](LICENSE)

# Nokotime MCP Server

A Model Context Protocol (MCP) server written in Go that integrates with [Nokotime](https://nokotime.com/) using [`github.com/gertv/go-noko`](https://github.com/gertv/go-noko) and [`github.com/mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go).

## Requirements

- Go 1.22+
- `NOKOTIME_API_KEY` environment variable set with your Nokotime Personal Access Token / API Key.

## Tools

### `list_projects`
Lists projects from Nokotime with pagination and optional filtering.

**Parameters:**
- `name` *(string, optional)*: Filter projects by project name.
- `billable` *(boolean, optional)*: Filter by billable status (`true` / `false`).
- `enabled` *(boolean, optional)*: Filter by active/enabled status (`true` / `false`).
- `project_group_ids` *(string, optional)*: Comma-separated list of project group IDs.

### `create_entry`
Creates a new timesheet entry in Nokotime.

**Parameters:**
- `date` *(string, required)*: Date of the entry in `YYYY-MM-DD` format (e.g. `2026-10-12`).
- `minutes` *(integer, required)*: Duration of the entry in minutes (e.g. `60` for 1 hour).
- `project_id` *(integer, optional)*: ID of the project for this entry.
- `project_name` *(string, optional)*: Name of the project (used if `project_id` is not provided).
- `description` *(string, optional)*: Description and hashtags for the entry (e.g. `#development Working on MCP`).
- `user_id` *(integer, optional)*: User ID to create the entry for.
- `user_email` *(string, optional)*: User email to create the entry for.
- `source_url` *(string, optional)*: URL associated with the entry.
- `billable` *(boolean, optional)*: Override billable status (`true` / `false`).

### `list_entries`
Lists timesheet entries from Nokotime with pagination and optional filtering.

**Parameters:**
- `from` *(string, optional)*: Start date filter in `YYYY-MM-DD` format.
- `to` *(string, optional)*: End date filter in `YYYY-MM-DD format.
- `project_ids` *(string, optional)*: Comma-separated list of project IDs to filter by.
- `user_ids` *(string, optional)*: Comma-separated list of user IDs to filter by.
- `tag_ids` *(string, optional)*: Comma-separated list of tag IDs to filter by.
- `tag_names` *(string, optional)*: Comma-separated list of tag names to filter by.
- `billable` *(boolean, optional)*: Filter by billable status (`true` / `false`).
- `invoiced` *(boolean, optional)*: Filter by invoiced status (`true` / `false`).
- `description` *(string, optional)*: Filter/search term for entry description.

## Build & Run

### Build Binary
```bash
go build -o server .
```

### Run Tests
```bash
go test -v ./...
```

### Environment Variable
Set the API key in your environment or `.envrc`:
```bash
export NOKOTIME_API_KEY="your-api-key-here"
```

### Test with MCP Inspector
You can test the server interactively using the official MCP Inspector:
```bash
npx @modelcontextprotocol/inspector ./server
```

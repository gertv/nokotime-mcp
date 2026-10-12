package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gertv/go-noko"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ListProjectsArgs struct {
	Name            string `json:"name,omitempty"`
	Billable        *bool  `json:"billable,omitempty"`
	Enabled         *bool  `json:"enabled,omitempty"`
	ProjectGroupIDs string `json:"project_group_ids,omitempty"`
}

type CreateEntryArgs struct {
	Date        string `json:"date"`
	Minutes     int    `json:"minutes"`
	ProjectID   *int   `json:"project_id,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
	Description string `json:"description,omitempty"`
	UserID      *int   `json:"user_id,omitempty"`
	UserEmail   string `json:"user_email,omitempty"`
	SourceURL   string `json:"source_url,omitempty"`
	Billable    *bool  `json:"billable,omitempty"`
}

type ListEntriesArgs struct {
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
	ProjectIDs  string `json:"project_ids,omitempty"`
	UserIDs     string `json:"user_ids,omitempty"`
	TagIDs      string `json:"tag_ids,omitempty"`
	TagNames    string `json:"tag_names,omitempty"`
	Billable    *bool  `json:"billable,omitempty"`
	Invoiced    *bool  `json:"invoiced,omitempty"`
	Description string `json:"description,omitempty"`
}

func main() {
	// Initialize MCP server instance
	s := server.NewMCPServer(
		"nokotime-mcp",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	// Define 'list_projects' tool
	listProjectsTool := mcp.NewTool("list_projects",
		mcp.WithDescription("List projects in Nokotime"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("name",
			mcp.Description("Optional filter by project name"),
		),
		mcp.WithBoolean("billable",
			mcp.Description("Optional filter by billable status (true or false)"),
		),
		mcp.WithBoolean("enabled",
			mcp.Description("Optional filter by enabled/active status (true or false)"),
		),
		mcp.WithString("project_group_ids",
			mcp.Description("Optional comma-separated list of project group IDs to filter by"),
		),
	)

	// Define 'create_entry' tool
	createEntryTool := mcp.NewTool("create_entry",
		mcp.WithDescription("Create a timesheet entry in Nokotime"),
		mcp.WithString("date",
			mcp.Required(),
			mcp.Description("Date of the entry in YYYY-MM-DD format (e.g. 2026-10-12)"),
		),
		mcp.WithInteger("minutes",
			mcp.Required(),
			mcp.Description("Duration of the entry in minutes (e.g. 60 for 1 hour)"),
		),
		mcp.WithInteger("project_id",
			mcp.Description("Optional ID of the project for this entry"),
		),
		mcp.WithString("project_name",
			mcp.Description("Optional name of the project (used if project_id is not provided)"),
		),
		mcp.WithString("description",
			mcp.Description("Optional description and hashtags for the entry (e.g. '#planning Team sync')"),
		),
		mcp.WithInteger("user_id",
			mcp.Description("Optional user ID to create the entry for"),
		),
		mcp.WithString("user_email",
			mcp.Description("Optional user email to create the entry for"),
		),
		mcp.WithString("source_url",
			mcp.Description("Optional URL associated with the entry"),
		),
		mcp.WithBoolean("billable",
			mcp.Description("Optional override for billable status"),
		),
	)

	// Define 'list_entries' tool
	listEntriesTool := mcp.NewTool("list_entries",
		mcp.WithDescription("List timesheet entries in Nokotime"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("from",
			mcp.Description("Optional start date filter in YYYY-MM-DD format"),
		),
		mcp.WithString("to",
			mcp.Description("Optional end date filter in YYYY-MM-DD format"),
		),
		mcp.WithString("project_ids",
			mcp.Description("Optional comma-separated list of project IDs to filter by"),
		),
		mcp.WithString("user_ids",
			mcp.Description("Optional comma-separated list of user IDs to filter by"),
		),
		mcp.WithString("tag_ids",
			mcp.Description("Optional comma-separated list of tag IDs to filter by"),
		),
		mcp.WithString("tag_names",
			mcp.Description("Optional comma-separated list of tag names to filter by"),
		),
		mcp.WithBoolean("billable",
			mcp.Description("Optional filter by billable status (true or false)"),
		),
		mcp.WithBoolean("invoiced",
			mcp.Description("Optional filter by invoiced status (true or false)"),
		),
		mcp.WithString("description",
			mcp.Description("Optional filter/search term for description"),
		),
	)

	// Register tool handlers
	s.AddTool(listProjectsTool, listProjectsHandler)
	s.AddTool(createEntryTool, createEntryHandler)
	s.AddTool(listEntriesTool, listEntriesHandler)

	// Start MCP server over stdio
	// Note: stdio transport uses stdout for JSON-RPC messages.
	// Any logging should be directed to stderr.
	log.SetOutput(os.Stderr)
	log.Println("Starting Nokotime MCP server on stdio...")

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func listProjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiKey := os.Getenv("NOKOTIME_API_KEY")
	if apiKey == "" {
		return mcp.NewToolResultError("NOKOTIME_API_KEY environment variable is not set"), nil
	}

	var args ListProjectsArgs
	if err := request.BindArguments(&args); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}

	client := noko.Nokotime("nokotime-mcp", apiKey)
	return fetchProjects(client, args)
}

func fetchProjects(client noko.Noko, args ListProjectsArgs) (*mcp.CallToolResult, error) {
	var paramSetters []noko.ParameterSetter

	if args.Name != "" {
		paramSetters = append(paramSetters, func(p noko.Parameters) {
			p["name"] = args.Name
		})
	}
	if args.Billable != nil {
		paramSetters = append(paramSetters, func(p noko.Parameters) {
			p["billable"] = strconv.FormatBool(*args.Billable)
		})
	}
	if args.Enabled != nil {
		paramSetters = append(paramSetters, func(p noko.Parameters) {
			p["enabled"] = strconv.FormatBool(*args.Enabled)
		})
	}
	if args.ProjectGroupIDs != "" {
		paramSetters = append(paramSetters, func(p noko.Parameters) {
			p["project_group_ids"] = args.ProjectGroupIDs
		})
	}

	page, err := client.ProjectsAPI().ListProjects(paramSetters...)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list projects: %v", err)), nil
	}

	var projects []noko.Project
	for proj := range page.AllProjects() {
		projects = append(projects, proj)
	}

	if projects == nil {
		projects = []noko.Project{}
	}

	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to serialize projects to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

func createEntryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiKey := os.Getenv("NOKOTIME_API_KEY")
	if apiKey == "" {
		return mcp.NewToolResultError("NOKOTIME_API_KEY environment variable is not set"), nil
	}

	var args CreateEntryArgs
	if err := request.BindArguments(&args); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}

	client := noko.Nokotime("nokotime-mcp", apiKey)
	return createEntry(client, args)
}

func createEntry(client noko.Noko, args CreateEntryArgs) (*mcp.CallToolResult, error) {
	if args.Date == "" {
		return mcp.NewToolResultError("date is required (format: YYYY-MM-DD)"), nil
	}
	if args.Minutes <= 0 {
		return mcp.NewToolResultError("minutes must be greater than 0"), nil
	}

	var inputSetters []noko.InputSetter

	if args.ProjectID != nil {
		inputSetters = append(inputSetters, func(i noko.Inputs) {
			i["project_id"] = *args.ProjectID
		})
	}
	if args.ProjectName != "" {
		inputSetters = append(inputSetters, func(i noko.Inputs) {
			i["project_name"] = args.ProjectName
		})
	}
	if args.Description != "" {
		inputSetters = append(inputSetters, func(i noko.Inputs) {
			i["description"] = args.Description
		})
	}
	if args.UserID != nil {
		inputSetters = append(inputSetters, func(i noko.Inputs) {
			i["user_id"] = *args.UserID
		})
	}
	if args.UserEmail != "" {
		inputSetters = append(inputSetters, func(i noko.Inputs) {
			i["user_email"] = args.UserEmail
		})
	}
	if args.SourceURL != "" {
		inputSetters = append(inputSetters, func(i noko.Inputs) {
			i["source_url"] = args.SourceURL
		})
	}
	if args.Billable != nil {
		inputSetters = append(inputSetters, func(i noko.Inputs) {
			i["billable"] = *args.Billable
		})
	}

	entry, err := client.EntriesAPI().CreateEntry(args.Date, args.Minutes, inputSetters...)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create entry: %v", err)), nil
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to serialize entry to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

func listEntriesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiKey := os.Getenv("NOKOTIME_API_KEY")
	if apiKey == "" {
		return mcp.NewToolResultError("NOKOTIME_API_KEY environment variable is not set"), nil
	}

	var args ListEntriesArgs
	if err := request.BindArguments(&args); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}

	client := noko.Nokotime("nokotime-mcp", apiKey)
	return fetchEntries(client, args)
}

func fetchEntries(client noko.Noko, args ListEntriesArgs) (*mcp.CallToolResult, error) {
	var paramSetters []noko.ParameterSetter

	if args.From != "" {
		paramSetters = append(paramSetters, func(p noko.Parameters) {
			p["from"] = args.From
		})
	}
	if args.To != "" {
		paramSetters = append(paramSetters, func(p noko.Parameters) {
			p["to"] = args.To
		})
	}
	if args.ProjectIDs != "" {
		paramSetters = append(paramSetters, func(p noko.Parameters) {
			p["project_ids"] = args.ProjectIDs
		})
	}
	if args.UserIDs != "" {
		paramSetters = append(paramSetters, func(p noko.Parameters) {
			p["user_ids"] = args.UserIDs
		})
	}
	if args.TagIDs != "" {
		paramSetters = append(paramSetters, func(p noko.Parameters) {
			p["tag_ids"] = args.TagIDs
		})
	}
	if args.TagNames != "" {
		paramSetters = append(paramSetters, func(p noko.Parameters) {
			p["tag_names"] = args.TagNames
		})
	}
	if args.Billable != nil {
		paramSetters = append(paramSetters, func(p noko.Parameters) {
			p["billable"] = strconv.FormatBool(*args.Billable)
		})
	}
	if args.Invoiced != nil {
		paramSetters = append(paramSetters, func(p noko.Parameters) {
			p["invoiced"] = strconv.FormatBool(*args.Invoiced)
		})
	}
	if args.Description != "" {
		paramSetters = append(paramSetters, func(p noko.Parameters) {
			p["description"] = args.Description
		})
	}

	page, err := client.EntriesAPI().ListEntries(paramSetters...)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list entries: %v", err)), nil
	}

	var entries []noko.Entry
	for entry := range page.AllEntries() {
		entries = append(entries, entry)
	}

	if entries == nil {
		entries = []noko.Entry{}
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to serialize entries to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

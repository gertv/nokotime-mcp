package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/gertv/go-noko"
	"github.com/mark3labs/mcp-go/mcp"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func mockHTTPClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestListProjectsHandlerMissingEnv(t *testing.T) {
	orig := os.Getenv("NOKOTIME_API_KEY")
	defer os.Setenv("NOKOTIME_API_KEY", orig)
	os.Unsetenv("NOKOTIME_API_KEY")

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}

	result, err := listProjectsHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Fatalf("expected error result when NOKOTIME_API_KEY is unset")
	}

	if len(result.Content) == 0 {
		t.Fatalf("expected content in result")
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("expected text content")
	}

	if textContent.Text != "NOKOTIME_API_KEY environment variable is not set" {
		t.Errorf("unexpected error message: %s", textContent.Text)
	}
}

func TestFetchProjectsSuccess(t *testing.T) {
	mockResponseJSON := `[
		{
			"id": 101,
			"name": "Project Alpha",
			"billable": true,
			"enabled": true
		},
		{
			"id": 102,
			"name": "Project Beta",
			"billable": false,
			"enabled": true
		}
	]`

	client := noko.Nokotime("nokotime-mcp", "test-token")
	client.Client(mockHTTPClient(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("X-FreckleToken") != "test-token" {
			t.Errorf("expected X-FreckleToken header test-token, got %s", req.Header.Get("X-FreckleToken"))
		}
		if req.URL.Path != "/v2/projects" {
			t.Errorf("expected path /v2/projects, got %s", req.URL.Path)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(mockResponseJSON)),
		}, nil
	}))

	result, err := fetchProjects(client, ListProjectsArgs{})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error result")
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("expected TextContent")
	}

	var projects []noko.Project
	if err := json.Unmarshal([]byte(textContent.Text), &projects); err != nil {
		t.Fatalf("failed to parse returned JSON: %v", err)
	}

	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].Name != "Project Alpha" || projects[0].Id != 101 {
		t.Errorf("unexpected first project: %+v", projects[0])
	}
	if projects[1].Name != "Project Beta" || projects[1].Id != 102 {
		t.Errorf("unexpected second project: %+v", projects[1])
	}
}

func TestFetchProjectsWithFilters(t *testing.T) {
	billable := true
	enabled := false
	args := ListProjectsArgs{
		Name:            "Alpha",
		Billable:        &billable,
		Enabled:         &enabled,
		ProjectGroupIDs: "12,34",
	}

	client := noko.Nokotime("nokotime-mcp", "test-token")
	client.Client(mockHTTPClient(func(req *http.Request) (*http.Response, error) {
		q := req.URL.Query()
		if q.Get("name") != "Alpha" {
			t.Errorf("expected query name=Alpha, got %s", q.Get("name"))
		}
		if q.Get("billable") != "true" {
			t.Errorf("expected query billable=true, got %s", q.Get("billable"))
		}
		if q.Get("enabled") != "false" {
			t.Errorf("expected query enabled=false, got %s", q.Get("enabled"))
		}
		if q.Get("project_group_ids") != "12,34" {
			t.Errorf("expected query project_group_ids=12,34, got %s", q.Get("project_group_ids"))
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString("[]")),
		}, nil
	}))

	result, err := fetchProjects(client, args)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error result")
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("expected TextContent")
	}
	if textContent.Text != "[]" {
		t.Errorf("expected empty array '[]', got %q", textContent.Text)
	}
}

func TestFetchProjectsAPIError(t *testing.T) {
	client := noko.Nokotime("nokotime-mcp", "invalid-token")
	client.Client(mockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Unauthorized"}`)),
		}, nil
	}))

	result, err := fetchProjects(client, ListProjectsArgs{})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error result on API error")
	}
}

func TestCreateEntryHandlerMissingEnv(t *testing.T) {
	orig := os.Getenv("NOKOTIME_API_KEY")
	defer os.Setenv("NOKOTIME_API_KEY", orig)
	os.Unsetenv("NOKOTIME_API_KEY")

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}

	result, err := createEntryHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error result when NOKOTIME_API_KEY is unset")
	}
}

func TestCreateEntryValidation(t *testing.T) {
	client := noko.Nokotime("nokotime-mcp", "test-token")

	// Missing date
	res, err := createEntry(client, CreateEntryArgs{Date: "", Minutes: 60})
	if err != nil || !res.IsError {
		t.Fatalf("expected error result when date is empty")
	}

	// Invalid minutes
	res, err = createEntry(client, CreateEntryArgs{Date: "2026-10-12", Minutes: 0})
	if err != nil || !res.IsError {
		t.Fatalf("expected error result when minutes is 0")
	}
}

func TestCreateEntrySuccess(t *testing.T) {
	mockResponseJSON := `{
		"id": 9999,
		"date": "2026-10-12",
		"minutes": 60,
		"description": "Internal work",
		"billable": false,
		"project": {
			"id": 311510,
			"name": "Internal"
		}
	}`

	projectID := 311510
	billable := false
	userID := 35643
	args := CreateEntryArgs{
		Date:        "2026-10-12",
		Minutes:     60,
		ProjectID:   &projectID,
		ProjectName: "Internal",
		Description: "Internal work",
		UserID:      &userID,
		UserEmail:   "gertv@anova.be",
		SourceURL:   "https://example.com",
		Billable:    &billable,
	}

	client := noko.Nokotime("nokotime-mcp", "test-token")
	client.Client(mockHTTPClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", req.Method)
		}
		if req.URL.Path != "/v2/entries" {
			t.Errorf("expected path /v2/entries, got %s", req.URL.Path)
		}

		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		var payload map[string]any
		if err := json.Unmarshal(bodyBytes, &payload); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		if payload["date"] != "2026-10-12" {
			t.Errorf("expected date 2026-10-12, got %v", payload["date"])
		}
		if payload["minutes"] != float64(60) {
			t.Errorf("expected minutes 60, got %v", payload["minutes"])
		}
		if payload["project_id"] != float64(311510) {
			t.Errorf("expected project_id 311510, got %v", payload["project_id"])
		}
		if payload["description"] != "Internal work" {
			t.Errorf("expected description 'Internal work', got %v", payload["description"])
		}

		return &http.Response{
			StatusCode: http.StatusCreated,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(mockResponseJSON)),
		}, nil
	}))

	result, err := createEntry(client, args)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error result")
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("expected TextContent")
	}

	var entry noko.Entry
	if err := json.Unmarshal([]byte(textContent.Text), &entry); err != nil {
		t.Fatalf("failed to unmarshal result JSON: %v", err)
	}
	if entry.Id != 9999 || entry.Minutes != 60 || entry.Date != "2026-10-12" {
		t.Errorf("unexpected entry returned: %+v", entry)
	}
}

func TestListEntriesHandlerMissingEnv(t *testing.T) {
	orig := os.Getenv("NOKOTIME_API_KEY")
	defer os.Setenv("NOKOTIME_API_KEY", orig)
	os.Unsetenv("NOKOTIME_API_KEY")

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}

	result, err := listEntriesHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error result when NOKOTIME_API_KEY is unset")
	}
}

func TestFetchEntriesSuccess(t *testing.T) {
	mockResponseJSON := `[
		{
			"id": 501,
			"date": "2026-10-12",
			"minutes": 60,
			"description": "Task 1",
			"billable": true,
			"project": {
				"id": 101,
				"name": "Project Alpha"
			}
		}
	]`

	client := noko.Nokotime("nokotime-mcp", "test-token")
	client.Client(mockHTTPClient(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/v2/entries" {
			t.Errorf("expected path /v2/entries, got %s", req.URL.Path)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(mockResponseJSON)),
		}, nil
	}))

	result, err := fetchEntries(client, ListEntriesArgs{})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error result")
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("expected TextContent")
	}

	var entries []noko.Entry
	if err := json.Unmarshal([]byte(textContent.Text), &entries); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Id != 501 || entries[0].Minutes != 60 {
		t.Errorf("unexpected entry: %+v", entries[0])
	}
}

func TestFetchEntriesWithFilters(t *testing.T) {
	billable := false
	invoiced := true
	args := ListEntriesArgs{
		From:        "2026-10-01",
		To:          "2026-10-12",
		ProjectIDs:  "311510",
		UserIDs:     "35643",
		TagIDs:      "1,2",
		TagNames:    "dev,test",
		Billable:    &billable,
		Invoiced:    &invoiced,
		Description: "Internal",
	}

	client := noko.Nokotime("nokotime-mcp", "test-token")
	client.Client(mockHTTPClient(func(req *http.Request) (*http.Response, error) {
		q := req.URL.Query()
		if q.Get("from") != "2026-10-01" {
			t.Errorf("expected query from=2026-10-01, got %s", q.Get("from"))
		}
		if q.Get("to") != "2026-10-12" {
			t.Errorf("expected query to=2026-10-12, got %s", q.Get("to"))
		}
		if q.Get("project_ids") != "311510" {
			t.Errorf("expected query project_ids=311510, got %s", q.Get("project_ids"))
		}
		if q.Get("user_ids") != "35643" {
			t.Errorf("expected query user_ids=35643, got %s", q.Get("user_ids"))
		}
		if q.Get("tag_ids") != "1,2" {
			t.Errorf("expected query tag_ids=1,2, got %s", q.Get("tag_ids"))
		}
		if q.Get("tag_names") != "dev,test" {
			t.Errorf("expected query tag_names=dev,test, got %s", q.Get("tag_names"))
		}
		if q.Get("billable") != "false" {
			t.Errorf("expected query billable=false, got %s", q.Get("billable"))
		}
		if q.Get("invoiced") != "true" {
			t.Errorf("expected query invoiced=true, got %s", q.Get("invoiced"))
		}
		if q.Get("description") != "Internal" {
			t.Errorf("expected query description=Internal, got %s", q.Get("description"))
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString("[]")),
		}, nil
	}))

	result, err := fetchEntries(client, args)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error result")
	}
}

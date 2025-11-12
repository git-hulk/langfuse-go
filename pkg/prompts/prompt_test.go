package prompts

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestListParams_ToQueryString(t *testing.T) {
	tests := []struct {
		name   string
		params ListParams
		want   string
	}{
		{"all params",
			ListParams{Name: "test", Label: "prod", Tag: "v1", Page: 1, Limit: 10},
			"name=test&label=prod&tag=v1&page=1&limit=10"},
		{"some params", ListParams{Name: "test", Page: 1}, "name=test&page=1"},
		{"no params", ListParams{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.params.ToQueryString())
		})
	}
}

func TestPromptClient_Get(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v2/prompts/test-prompt", r.URL.Path)
			prompt := PromptEntry{Name: "test-prompt"}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(prompt)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewClient(cli)
	prompt, err := client.Get(context.Background(), GetParams{Name: "test-prompt"})
	require.NoError(t, err)
	require.Equal(t, "test-prompt", prompt.Name)
}

func TestPromptClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v2/prompts", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"meta":{"page":1,"limit":10,"totalItems":1,"totalPages":1},"data":[{"name":"test-prompt"}]}`))
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewClient(cli)
	promptList, err := client.List(context.Background(), ListParams{})
	require.NoError(t, err)
	require.Len(t, promptList.Data, 1)
	require.Equal(t, "test-prompt", promptList.Data[0].Name)
	// verify meta
	require.Equal(t, 1, promptList.Metadata.Page)
	require.Equal(t, 10, promptList.Metadata.Limit)
	require.Equal(t, 1, promptList.Metadata.TotalItems)
	require.Equal(t, 1, promptList.Metadata.TotalPages)
}

func TestPromptClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v2/prompts", r.URL.Path)
			var prompt PromptEntry
			err := json.NewDecoder(r.Body).Decode(&prompt)
			require.NoError(t, err)
			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(prompt)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewClient(cli)
	createPrompt := &PromptEntry{Name: "test-prompt", Prompt: []ChatMessageWithPlaceHolder{{Role: "user", Content: "hello"}}}
	prompt, err := client.Create(context.Background(), createPrompt)
	require.NoError(t, err)
	require.Equal(t, "test-prompt", prompt.Name)
}

func TestPromptEntryCompile_Text(t *testing.T) {
	entry := &PromptEntry{
		Name:   "text",
		Type:   "text",
		Prompt: "Hello {{ name }}!",
	}

	result, err := entry.Compile(map[string]any{"name": "Alice"})
	require.NoError(t, err)
	require.Equal(t, "Hello Alice!", result)

	resultNoVars, err := entry.Compile(map[string]any{})
	require.NoError(t, err)
	require.Equal(t, "Hello {{ name }}!", resultNoVars)
}

func TestPromptEntryCompile_Chat(t *testing.T) {
	entry := &PromptEntry{
		Name: "chat",
		Type: "chat",
		Prompt: []ChatMessageWithPlaceHolder{
			{Role: "system", Content: "Hi {{ user }}", Type: ChatMessageTypeMessage},
			{Name: "examples", Type: ChatMessageTypePlaceHolder},
		},
	}

	vars := map[string]any{
		"user": "Bob",
		"examples": []ChatMessageWithPlaceHolder{
			{Role: "user", Content: "Example {{ user }}", Type: ChatMessageTypeMessage},
		},
	}

	result, err := entry.Compile(vars)
	require.NoError(t, err)

	compiled, ok := result.([]ChatMessageWithPlaceHolder)
	require.True(t, ok)
	require.Len(t, compiled, 2)
	require.Equal(t, "Hi Bob", compiled[0].Content)
	require.Equal(t, "Example Bob", compiled[1].Content)
}

func TestPromptEntryCompile_MissingPlaceholderVariable(t *testing.T) {
	entry := &PromptEntry{
		Name: "chat",
		Type: "chat",
		Prompt: []ChatMessageWithPlaceHolder{
			{Name: "examples", Type: ChatMessageTypePlaceHolder},
		},
	}

	_, err := entry.Compile(map[string]any{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing variable for placeholder 'examples'")
}

func TestPromptEntryCompile_PlaceholderWithStringValue(t *testing.T) {
	entry := &PromptEntry{
		Name: "chat",
		Type: "chat",
		Prompt: []ChatMessageWithPlaceHolder{
			{Name: "examples", Type: ChatMessageTypePlaceHolder},
		},
	}

	result, err := entry.Compile(map[string]any{"examples": "inline"})
	require.NoError(t, err)

	compiled := result.([]ChatMessageWithPlaceHolder)
	require.Len(t, compiled, 1)
	require.Equal(t, "inline", compiled[0].Content)
	require.Equal(t, ChatMessageTypePlaceHolder, compiled[0].Type)
}

func TestPromptEntryCompile_NestedPlaceholderReturnsError(t *testing.T) {
	entry := &PromptEntry{
		Name: "chat",
		Type: "chat",
		Prompt: []ChatMessageWithPlaceHolder{
			{Name: "examples", Type: ChatMessageTypePlaceHolder},
		},
	}

	vars := map[string]any{
		"examples": []ChatMessageWithPlaceHolder{
			{Name: "inner", Type: ChatMessageTypePlaceHolder},
		},
	}

	_, err := entry.Compile(vars)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nested placeholders are not allowed")
}

func TestPromptEntryCompile_NilPromptEntry(t *testing.T) {
	var entry *PromptEntry
	_, err := entry.Compile(map[string]any{})
	require.EqualError(t, err, "prompt entry is empty")
}

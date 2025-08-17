package prompt

import (
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPromptClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/prompts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		prompt := Prompt{
			Name:              "test-prompt",
			Version:           1,
			Prompt:            "This is a test prompt",
			IsActive:          true,
			LangfuseCreatedAt: time.Now(),
			LangfuseUpdatedAt: time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(prompt); err != nil {
			t.Errorf("failed to encode prompt: %s", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	promptClient := NewPromptClient(cli)
	prompt, err := promptClient.Get("test-prompt", 1, "test-label")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if prompt == nil {
		t.Errorf("prompt is nil")
	}
	if prompt.Name != "test-prompt" {
		t.Errorf("unexpected prompt name: %s", prompt.Name)
	}
}

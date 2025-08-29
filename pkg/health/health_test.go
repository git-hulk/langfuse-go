package health

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/go-resty/resty/v2"
    "github.com/stretchr/testify/require"
)

func TestHealthClient_Check_Success(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        require.Equal(t, "/health", r.URL.Path)
        require.Equal(t, http.MethodGet, r.Method)

        resp := HealthResponse{Version: "1.2.3", Status: "OK"}
        w.Header().Set("Content-Type", "application/json")
        err := json.NewEncoder(w).Encode(resp)
        require.NoError(t, err)
    }))
    defer server.Close()

    cli := resty.New().SetBaseURL(server.URL)
    client := NewClient(cli)

    health, err := client.Check(context.Background())
    require.NoError(t, err)
    require.NotNil(t, health)
    require.Equal(t, "1.2.3", health.Version)
    require.Equal(t, "OK", health.Status)
}

func TestHealthClient_Check_HTTPError(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        require.Equal(t, "/health", r.URL.Path)
        w.WriteHeader(http.StatusInternalServerError)
    }))
    defer server.Close()

    cli := resty.New().SetBaseURL(server.URL)
    client := NewClient(cli)

    health, err := client.Check(context.Background())
    require.Error(t, err)
    require.Nil(t, health)
    require.Contains(t, err.Error(), "get health failed with status code 500")
}

func TestHealthClient_Check_TransportError(t *testing.T) {
    // resty client without BaseURL will fail for relative path "/health"
    cli := resty.New()
    client := NewClient(cli)

    health, err := client.Check(context.Background())
    require.Error(t, err)
    require.Nil(t, health)
}


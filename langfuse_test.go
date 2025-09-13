package langfuse

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewClient_WithoutOptions(t *testing.T) {
	client := NewClient("https://cloud.langfuse.com", "public-key", "secret-key")

	require.NotNil(t, client)
	require.NotNil(t, client.restyCli)
	require.NotNil(t, client.ingestor)
	require.NotNil(t, client.prompt)
	require.NotNil(t, client.model)
	require.NotNil(t, client.project)
	require.NotNil(t, client.comment)
	require.NotNil(t, client.dataset)
	require.NotNil(t, client.session)
	require.NotNil(t, client.score)
	require.NotNil(t, client.llmConnection)
	require.NotNil(t, client.organization)
	require.NotNil(t, client.health)
	require.NotNil(t, client.media)
}

func TestNewClient_WithHTTPClient(t *testing.T) {
	customHTTPClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
		},
	}

	client := NewClient("https://cloud.langfuse.com", "public-key", "secret-key", WithHTTPClient(customHTTPClient))

	require.NotNil(t, client)
	require.NotNil(t, client.restyCli)

	// Verify that the custom HTTP client is being used
	restyHTTPClient := client.restyCli.GetClient()
	require.Equal(t, customHTTPClient, restyHTTPClient)
	require.Equal(t, 30*time.Second, restyHTTPClient.Timeout)

	// Verify that all subclients are properly initialized
	require.NotNil(t, client.ingestor)
	require.NotNil(t, client.prompt)
	require.NotNil(t, client.model)
	require.NotNil(t, client.project)
	require.NotNil(t, client.comment)
	require.NotNil(t, client.dataset)
	require.NotNil(t, client.session)
	require.NotNil(t, client.score)
	require.NotNil(t, client.llmConnection)
	require.NotNil(t, client.organization)
	require.NotNil(t, client.health)
	require.NotNil(t, client.media)
}

func TestNewClient_WithMultipleOptions(t *testing.T) {
	customHTTPClient := &http.Client{
		Timeout: 45 * time.Second,
	}

	client := NewClient("https://cloud.langfuse.com", "public-key", "secret-key", WithHTTPClient(customHTTPClient))

	require.NotNil(t, client)

	// Verify that the custom HTTP client is being used
	restyHTTPClient := client.restyCli.GetClient()
	require.Equal(t, customHTTPClient, restyHTTPClient)
	require.Equal(t, 45*time.Second, restyHTTPClient.Timeout)
}

func TestWithHTTPClient(t *testing.T) {
	customHTTPClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	config := &clientConfig{}
	option := WithHTTPClient(customHTTPClient)
	option(config)

	require.Equal(t, customHTTPClient, config.httpClient)
}

func TestClientConfig_Default(t *testing.T) {
	config := &clientConfig{}
	require.Nil(t, config.httpClient)
}

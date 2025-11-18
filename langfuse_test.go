package langfuse

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/git-hulk/langfuse-go/pkg/traces"
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

func TestTrace(t *testing.T) {
	// Use test environment configuration instead of real environment sensitive information
	client := NewClient("http://localhost:3000", "test-public-key", "test-secret-key")

	// Create a trace
	trace := client.StartTrace(context.Background(), "Test Trace")
	trace.Input = map[string]string{"input": "Test input"}
	trace.Output = map[string]string{"output": "Test output"}
	trace.Tags = []string{"test", "example"}

	// Test Agent type observation
	agent := trace.StartObservation("test_agent", traces.ObservationTypeAgent)
	agent.Input = map[string]string{"input": "Test agent input"}
	agent.Output = map[string]string{"output": "Test agent output"}

	// Test Retriever type observation
	retriever := trace.StartObservation("test_retriever", traces.ObservationTypeRetriever)
	retriever.Input = map[string]string{"input": "Test retriever input"}
	retriever.Output = map[string]string{"output": "Test retriever output"}
	retriever.End()

	// Test Generation type observation (LLM)
	llm := trace.StartGeneration("test_generation")
	llm.Input = map[string]string{"input": "Test generation input"}
	llm.Output = map[string]string{"output": "Test generation output"}
	llm.Usage = traces.Usage{
		Input:  10,
		Output: 20,
		Total:  30,
		Unit:   traces.UnitTokens,
	}
	llm.End()

	// Test Tool type observation
	tool := trace.StartObservation("test_tool", traces.ObservationTypeTool)
	tool.Input = map[string]string{"input": "Test tool input"}
	tool.Output = map[string]string{"output": "Test tool output"}
	tool.End()

	// End Agent observation
	agent.End()

	// End trace
	trace.End()

	// Flush client to ensure all data is processed
	client.Flush()

	// Verify trace was created correctly
	require.NotEmpty(t, trace.ID, "Trace ID should not be empty")
	require.Equal(t, "Test Trace", trace.Name, "Trace name should match")

	// Verify trace fields
	require.Equal(t, map[string]string{"input": "Test input"}, trace.Input, "Trace input should match")
	require.Equal(t, map[string]string{"output": "Test output"}, trace.Output, "Trace output should match")
	require.Equal(t, []string{"test", "example"}, trace.Tags, "Trace tags should match")

	// Verify each observation type was created with correct parameters
	// Agent observation
	require.Equal(t, "test_agent", agent.Name, "Agent name should match")
	require.Equal(t, traces.ObservationTypeAgent, agent.Type, "Agent type should be AGENT")
	require.Equal(t, map[string]string{"input": "Test agent input"}, agent.Input, "Agent input should match")
	require.Equal(t, map[string]string{"output": "Test agent output"}, agent.Output, "Agent output should match")

	// Retriever observation
	require.Equal(t, "test_retriever", retriever.Name, "Retriever name should match")
	require.Equal(t, traces.ObservationTypeRetriever, retriever.Type, "Retriever type should be RETRIEVER")
	require.Equal(t, map[string]string{"input": "Test retriever input"}, retriever.Input, "Retriever input should match")
	require.Equal(t, map[string]string{"output": "Test retriever output"}, retriever.Output, "Retriever output should match")

	// Generation observation
	require.Equal(t, "test_generation", llm.Name, "Generation name should match")
	require.Equal(t, traces.ObservationTypeGeneration, llm.Type, "Generation type should be GENERATION")
	require.Equal(t, map[string]string{"input": "Test generation input"}, llm.Input, "Generation input should match")
	require.Equal(t, map[string]string{"output": "Test generation output"}, llm.Output, "Generation output should match")
	require.Equal(t, 10, llm.Usage.Input, "Generation input usage should match")
	require.Equal(t, 20, llm.Usage.Output, "Generation output usage should match")
	require.Equal(t, 30, llm.Usage.Total, "Generation total usage should match")
	require.Equal(t, traces.UnitTokens, llm.Usage.Unit, "Generation usage unit should match")

	// Tool observation
	require.Equal(t, "test_tool", tool.Name, "Tool name should match")
	require.Equal(t, traces.ObservationTypeTool, tool.Type, "Tool type should be TOOL")
	require.Equal(t, map[string]string{"input": "Test tool input"}, tool.Input, "Tool input should match")
	require.Equal(t, map[string]string{"output": "Test tool output"}, tool.Output, "Tool output should match")

	// Verify observations ended correctly (end time should be set)
	require.NotNil(t, retriever.EndTime, "Retriever end time should be set")
	require.NotNil(t, llm.EndTime, "Generation end time should be set")
	require.NotNil(t, tool.EndTime, "Tool end time should be set")
	require.NotNil(t, agent.EndTime, "Agent end time should be set")

	// Verify observation durations are reasonable (greater than or equal to 0)
	retrieverDuration := retriever.EndTime.Sub(retriever.StartTime)
	require.True(t, retrieverDuration >= 0, "Retriever duration should be non-negative")

	llmDuration := llm.EndTime.Sub(llm.StartTime)
	require.True(t, llmDuration >= 0, "Generation duration should be non-negative")

	toolDuration := tool.EndTime.Sub(tool.StartTime)
	require.True(t, toolDuration >= 0, "Tool duration should be non-negative")

	agentDuration := agent.EndTime.Sub(agent.StartTime)
	require.True(t, agentDuration >= 0, "Agent duration should be non-negative")
}

// Package langfuse provides a Go client library for interacting with the Langfuse platform.
//
// This package offers comprehensive support for observability tracing, prompt management,
// model configuration, datasets, sessions, scores, projects, LLM connections, comments,
// and annotations functionality with efficient batch processing.
//
// Basic usage:
//
//	client := langfuse.NewClient("https://cloud.langfuse.com", "your-public-key", "your-secret-key")
//	defer client.Close()
//
//	trace := client.StartTrace("my-application")
//	span := trace.StartSpan("processing-step")
//	// ... your application logic
//	span.End()
//	trace.End()
package langfuse

import (
	"github.com/git-hulk/langfuse-go/pkg/organizations"
	"github.com/go-resty/resty/v2"

	"github.com/git-hulk/langfuse-go/pkg/comments"
	"github.com/git-hulk/langfuse-go/pkg/datasets"
	"github.com/git-hulk/langfuse-go/pkg/llmconnections"
	"github.com/git-hulk/langfuse-go/pkg/models"
	"github.com/git-hulk/langfuse-go/pkg/projects"
	"github.com/git-hulk/langfuse-go/pkg/prompts"
	"github.com/git-hulk/langfuse-go/pkg/scores"
	"github.com/git-hulk/langfuse-go/pkg/sessions"
	"github.com/git-hulk/langfuse-go/pkg/traces"
)

// LangFuse is the main client for interacting with the Langfuse platform.
//
// It provides access to all Langfuse functionality including tracing, prompts,
// models, datasets, sessions, scores, projects, LLM connections, comments,
// and annotations through dedicated client instances.
//
// The client manages HTTP connections and provides efficient batch processing
// for trace ingestion with automatic flushing and graceful shutdown capabilities.
type LangFuse struct {
	ingestor      *traces.Ingestor
	prompt        *prompts.Client
	model         *models.Client
	project       *projects.Client
	comment       *comments.Client
	dataset       *datasets.Client
	session       *sessions.Client
	score         *scores.Client
	llmConnection *llmconnections.Client
	organization  *organizations.Client
	restyCli      *resty.Client
}

// NewClient creates a new LangFuse client instance with the specified host and credentials.
//
// The host should be the base URL of your Langfuse instance (e.g., "https://cloud.langfuse.com").
// The publicKey and secretKey are obtained from your Langfuse project settings.
//
// The client automatically configures HTTP basic authentication and sets the API base URL.
// Remember to call Close() when done to ensure all pending traces are flushed.
func NewClient(host string, publicKey string, secretKey string) *LangFuse {
	restyCli := resty.New().
		SetBaseURL(host+"/api/public").
		SetBasicAuth(publicKey, secretKey)

	return &LangFuse{
		ingestor:      traces.NewIngestor(restyCli),
		prompt:        prompts.NewClient(restyCli),
		model:         models.NewClient(restyCli),
		project:       projects.NewClient(restyCli),
		comment:       comments.NewClient(restyCli),
		dataset:       datasets.NewClient(restyCli),
		session:       sessions.NewClient(restyCli),
		score:         scores.NewClient(restyCli),
		llmConnection: llmconnections.NewClient(restyCli),
		organization:  organizations.NewClient(restyCli),
		restyCli:      restyCli,
	}
}

// StartTrace creates a new trace with the given name.
//
// A trace represents a single execution flow in your application and can contain
// multiple observations (spans). Traces are automatically batched and sent to
// Langfuse for efficient ingestion.
//
// Returns a Trace instance that you can use to add observations and metadata.
func (c *LangFuse) StartTrace(name string) *traces.Trace {
	return c.ingestor.StartTrace(name)
}

// Prompts returns a client for managing prompt templates and versions.
//
// Use this client to create, retrieve, list, and manage prompt templates
// for your AI applications.
func (c *LangFuse) Prompts() *prompts.Client {
	return c.prompt
}

// Models returns a client for managing model configurations and pricing.
//
// Use this client to define model pricing, match patterns, and manage
// model metadata for cost tracking and analytics.
func (c *LangFuse) Models() *models.Client {
	return c.model
}

// Projects returns a client for managing projects and API keys.
//
// Use this client to create, update, and manage projects, as well as
// manage API keys within projects. Most operations require organization-scoped API keys.
func (c *LangFuse) Projects() *projects.Client {
	return c.project
}

// Comments returns a client for managing comments on traces, observations, and sessions.
//
// Use this client to add contextual comments to your traces and observations
// for collaboration and debugging purposes.
func (c *LangFuse) Comments() *comments.Client {
	return c.comment
}

// Datasets returns a client for managing datasets and dataset items.
//
// Use this client to create and manage datasets for training, evaluation,
// and testing of your AI models, including dataset items and runs.
func (c *LangFuse) Datasets() *datasets.Client {
	return c.dataset
}

// Sessions returns a client for managing user sessions and their associated traces.
//
// Use this client to retrieve and analyze user sessions, including
// filtering by time ranges and environments.
func (c *LangFuse) Sessions() *sessions.Client {
	return c.session
}

// Scores returns a client for managing evaluation scores and score configurations.
//
// Use this client to create, retrieve, and manage scores for your traces
// and observations, including score configurations for different data types.
func (c *LangFuse) Scores() *scores.Client {
	return c.score
}

// LLMConnections returns a client for managing LLM provider connections.
//
// Use this client to configure connections to various LLM providers
// like OpenAI, Anthropic, Azure OpenAI, AWS Bedrock, and Google Vertex AI.
func (c *LangFuse) LLMConnections() *llmconnections.Client {
	return c.llmConnection
}

// Organizations returns a client for managing organization and project memberships.
//
// Use this client to manage user roles and permissions within organizations
// and projects. Most operations require organization-scoped API keys.
func (c *LangFuse) Organizations() *organizations.Client {
	return c.organization
}

// Close gracefully shuts down the client and flushes all pending traces.
//
// This method ensures that all batched traces are sent to Langfuse before
// the client is closed. It should be called when you're done using the client,
// typically in a defer statement.
//
// Returns an error if the shutdown process fails or times out.
func (c *LangFuse) Close() error {
	return c.ingestor.Close()
}

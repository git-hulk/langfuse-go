# langfuse-go ![GitHub CI](https://github.com/git-hulk/langfuse-go/actions/workflows/ci.yaml/badge.svg) [![LICENSE](https://img.shields.io/github/license/git-hulk/langfuse-go.svg)](https://github.com/git-hulk/langfuse-go/blob/master/LICENSE) [![GoDoc](https://img.shields.io/badge/Godoc-reference-blue.svg)](https://godoc.org/github.com/git-hulk/langfuse-go) 

Go client & SDK for interacting with [Langfuse](https://langfuse.com/). Provides comprehensive support for observability tracing, prompt management, model configuration, datasets, sessions, scores, projects, LLM connections, and organization management with efficient batch processing.

## Table of Contents

- [Installation](#installation)
- [Features](#features)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
  - [Core Observability](#core-observability)
    - [Tracing](#tracing)
    - [Sessions](#sessions)
    - [Comments](#comments)
  - [AI/ML Management](#aiml-management)
    - [Prompts](#prompts)
    - [Models](#models)
    - [Scores](#scores)
    - [LLM Connections](#llm-connections)
  - [Data Management](#data-management)
    - [Datasets](#datasets)
  - [Organization & Projects](#organization--projects)
    - [Projects](#projects)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Quick Start

```go
package main

import (
    "context"
    
    langfuse "github.com/git-hulk/langfuse-go"
)

func main() {
    // Initialize the client
    client := langfuse.NewClient("YOUR_HOST", "YOUR_PUBLIC_KEY", "YOUR_PRIVATE_KEY")
    defer client.Close() // Ensure all pending traces are flushed

    // Start tracing
    trace := client.StartTrace("my-application")
    span := trace.StartSpan("processing-step")
    
    // Your application logic here...
    nestedSpan  := span.StartSpan("nested-processing") 
	nestedSpan.End()
	
    span.End()
    trace.End()
}
```

## API Reference

## Core Observability

Core functionality for tracking and monitoring your AI applications with distributed tracing, session management, and contextual comments.

### Tracing

```go
import (
    "context"

    langfuse "github.com/git-hulk/langfuse-go"
)

func main() {
    langfuse := langfuse.NewClient("YOUR_HOST", "YOUR_PUBLIC_KEY", "YOUR_PRIVATE_KEY")

    ctx := context.Background()
    trace := langfuse.StartTrace("it's a trace")
    span := trace.StartSpan("it's a span")
    span.End()
    trace.End()
    langfuse.Close() // flushes all pending traces
}
```

### Sessions

```go
import (
    "context"
    "time"

    langfuse "github.com/git-hulk/langfuse-go"
    "github.com/git-hulk/langfuse-go/pkg/sessions"
)

func main() {
    langfuse := langfuse.NewClient("YOUR_HOST", "YOUR_PUBLIC_KEY", "YOUR_PRIVATE_KEY")

    ctx := context.Background()

    // Get a session by ID with its traces
    session, err := langfuse.Sessions().Get(ctx, "session-123")

    // List sessions with filters
    sessionsList, err := langfuse.Sessions().List(ctx, sessions.ListParams{
        Page:          1,
        Limit:         10,
        FromTimestamp: time.Now().Add(-24 * time.Hour),
        ToTimestamp:   time.Now(),
        Environment:   []string{"production"},
    })
}
```

### Comments

```go
import (
    "context"

    langfuse "github.com/git-hulk/langfuse-go"
    "github.com/git-hulk/langfuse-go/pkg/comments"
)

func main() {
    langfuse := langfuse.NewClient("YOUR_HOST", "YOUR_PUBLIC_KEY", "YOUR_PRIVATE_KEY")

    ctx := context.Background()

    // Create a comment on a trace
    createdComment, err := langfuse.Comments().Create(ctx, &comments.CreateCommentRequest{
        ObjectType: comments.ObjectTypeTrace,
        ObjectID:   "trace-123",
        Content:    "This trace looks good!",
    })

    // Get a comment by ID
    comment, err := langfuse.Comments().Get(ctx, "comment-id")

    // List comments with filters
    commentsList, err := langfuse.Comments().List(ctx, comments.ListParams{
        ObjectType: comments.ObjectTypeTrace,
        ObjectID:   "trace-123",
        Page:       1,
        Limit:      10,
    })
}
```

## AI/ML Management

Tools for managing prompts, models, evaluation scores, and LLM provider connections for your AI applications.

### Prompts

```go
import (
    "context"

    langfuse "github.com/git-hulk/langfuse-go"
    "github.com/git-hulk/langfuse-go/pkg/prompts"
)

func main() {
    langfuse := langfuse.NewClient("YOUR_HOST", "YOUR_PUBLIC_KEY", "YOUR_PRIVATE_KEY")

    ctx := context.Background()

    createdPrompt, err := langfuse.Prompts().Create(ctx, prompts.PromptEntry{
        Name: "welcome-message",
        Prompt: []prompts.ChatMessageWithPlaceHolder {
            {Role: "system", Content: "You are a helpful assistant."},
            {Role: "user", Content: "Hello!"},
        }
	})

    prompt, err := langfuse.Prompts().Get(ctx, prompts.GetParams{Name: "welcome-message"})

    listResponse, err := langfuse.Prompts().List(ctx, prompts.ListParams{Limit: 20})
}
```

### Models

```go
import (
    "context"

    langfuse "github.com/git-hulk/langfuse-go"
    "github.com/git-hulk/langfuse-go/pkg/models"
)

func main() {
    langfuse := langfuse.NewClient("YOUR_HOST", "YOUR_PUBLIC_KEY", "YOUR_PRIVATE_KEY")

    ctx := context.Background()

    // Create a new model
    createdModel, err := langfuse.Models().Create(ctx, &models.ModelEntry{
        ModelName:    "gpt-4",
        MatchPattern: "gpt-4*",
        InputPrice:   0.03,
        OutputPrice:  0.06,
        Unit:         "TOKENS",
    })

    // Get a model by ID
    model, err := langfuse.Models().Get(ctx, "model-id")

    // List models
    listModels, err := langfuse.Models().List(ctx, models.ListParams{
        Page:  1,
        Limit: 20,
    })

    // Delete a model
    err = langfuse.Models().Delete(ctx, "model-id")
}
```

### Scores

```go
import (
    "context"

    langfuse "github.com/git-hulk/langfuse-go"
    "github.com/git-hulk/langfuse-go/pkg/scores"
)

func main() {
    langfuse := langfuse.NewClient("YOUR_HOST", "YOUR_PUBLIC_KEY", "YOUR_PRIVATE_KEY")

    ctx := context.Background()

    // Create a score for a trace
    createdScore, err := langfuse.Scores().Create(ctx, &scores.CreateScoreRequest{
        TraceID:  "trace-123",
        Name:     "accuracy",
        Value:    0.95,
        DataType: scores.ScoreDataTypeNumeric,
        Comment:  "High accuracy score",
    })

    // Get a score by ID
    score, err := langfuse.Scores().Get(ctx, "score-id")

    // List scores with filters
    scoresList, err := langfuse.Scores().List(ctx, scores.ListParams{
        Page:   1,
        Limit:  20,
        Name:   "accuracy",
        Source: scores.ScoreSourceAPI,
    })

    // Delete a score
    err = langfuse.Scores().Delete(ctx, "score-id")
}
```

### LLM Connections

```go
import (
    "context"

    langfuse "github.com/git-hulk/langfuse-go"
    "github.com/git-hulk/langfuse-go/pkg/llmconnections"
)

func main() {
    langfuse := langfuse.NewClient("YOUR_HOST", "YOUR_PUBLIC_KEY", "YOUR_PRIVATE_KEY")

    ctx := context.Background()

    // List LLM connections
    connections, err := langfuse.LLMConnections().List(ctx, llmconnections.ListParams{
        Page:  1,
        Limit: 10,
    })

    // Create or update an LLM connection
    connection, err := langfuse.LLMConnections().Upsert(ctx, &llmconnections.UpsertLLMConnectionRequest{
        Provider:          "OpenAI",
        Adapter:           llmconnections.AdapterOpenAI,
        SecretKey:         "sk-your-openai-key",
        CustomModels:      []string{"gpt-4", "gpt-3.5-turbo"},
        WithDefaultModels: &[]bool{true}[0],
        ExtraHeaders:      map[string]string{"Custom-Header": "value"},
    })
}
```

## Data Management

Manage datasets for training, evaluation, and testing of your AI models.

### Datasets

```go
import (
    "context"

    langfuse "github.com/git-hulk/langfuse-go"
    "github.com/git-hulk/langfuse-go/pkg/datasets"
)

func main() {
    langfuse := langfuse.NewClient("YOUR_HOST", "YOUR_PUBLIC_KEY", "YOUR_PRIVATE_KEY")

    ctx := context.Background()

    // Create a new dataset
    createdDataset, err := langfuse.Datasets().Create(ctx, &datasets.CreateDatasetRequest{
        Name:        "evaluation-dataset",
        Description: "Dataset for model evaluation",
        Metadata:    map[string]interface{}{"version": "1.0"},
    })

    // Get a dataset by name
    dataset, err := langfuse.Datasets().Get(ctx, "evaluation-dataset")

    // List datasets
    listDatasets, err := langfuse.Datasets().List(ctx, datasets.ListParams{
        Page:  1,
        Limit: 20,
    })
}
```

## Organization & Projects

Manage projects, API keys, and organization memberships. Most operations require organization-scoped API keys.

### Projects

```go
import (
    "context"

    langfuse "github.com/git-hulk/langfuse-go"
    "github.com/git-hulk/langfuse-go/pkg/projects"
)

func main() {
    langfuse := langfuse.NewClient("YOUR_HOST", "YOUR_PUBLIC_KEY", "YOUR_PRIVATE_KEY")

    ctx := context.Background()

    // Get projects associated with your API key
    projects, err := langfuse.Projects().Get(ctx)

    // Create a new project (requires organization-scoped API key)
    createdProject, err := langfuse.Projects().Create(ctx, &projects.CreateProjectRequest{
        Name:      "my-new-project",
        Metadata:  map[string]interface{}{"team": "ai"},
        Retention: 30,
    })

    // Update a project (requires organization-scoped API key)
    updatedProject, err := langfuse.Projects().Update(ctx, "project-id", &projects.UpdateProjectRequest{
        Name:      "updated-project-name",
        Retention: 60,
    })

    // Delete a project (requires organization-scoped API key)
    deleteResponse, err := langfuse.Projects().Delete(ctx, "project-id")

    // Manage API keys for a project (requires organization-scoped API key)
    apiKeys, err := langfuse.Projects().GetAPIKeys(ctx, "project-id")
    
    newAPIKey, err := langfuse.Projects().CreateAPIKey(ctx, "project-id", &projects.CreateAPIKeyRequest{
        Note: &[]string{"API key for production"}[0],
    })
    
    deleteAPIResponse, err := langfuse.Projects().DeleteAPIKey(ctx, "project-id", "api-key-id")
}
```

## Development

### Testing

```bash
make test                    # Run all tests with race detector (-race -count=1)
go test ./...               # Standard Go test runner  
go test ./pkg/datasets/     # Test specific package
go test -v ./pkg/traces/    # Verbose output for specific package
```

### Code Formatting

```bash
make format                 # Format with goimports + gofmt (includes local import ordering)
goimports -w -local github.com/git-hulk/langfuse-go ./...
```

### Build & Linting

```bash
go build ./...              # Build all packages
golangci-lint run           # Lint (CI uses v1.64.7)
```

## Contributing

Issues & PRs are welcome. Please include tests for new functionality or bug fixes.

## License

MIT License. See [LICENSE](LICENSE).

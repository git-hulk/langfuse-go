# langfuse-go

Go client & SDK for interacting with [langfuse](https://langfuse.com/): tracing, prompt management and ingestion batching.

## Features

- Trace 
- Prompt
- Model
- Comments 

And more to come...

## API Reference

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

## Run Tests 

```bash
make test    # run tests with race detector
```

## Format Code

```bash
make format 
```

## Contributing

Issues & PRs are welcome. Please include tests for new functionality or bug fixes.

## License

MIT License. See [LICENSE](LICENSE).

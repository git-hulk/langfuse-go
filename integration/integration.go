package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/git-hulk/langfuse-go"
	"github.com/git-hulk/langfuse-go/pkg/models"
)

func runTraceTests(client *langfuse.LangFuse) {
	trace := client.StartTrace("Test Trace")
	trace.Input = map[string]string{"input": "Test input"}
	trace.Output = map[string]string{"output": "Test output"}
	trace.Tags = []string{"test", "example"}

	// Start a span within the trace
	span := trace.StartSpan("Test Span")
	span.Input = map[string]string{"span_input": "Processing data..."}
	span.Output = map[string]string{"span_output": "Data processed successfully!"}
	span.End()

	trace.End()
}

func runModelTests(client *langfuse.LangFuse) {
	ctx := context.Background()
	modelClient := client.Models()

	fmt.Println("Testing Model API...")

	// Test creating a model
	testModel := &models.ModelEntry{
		ModelName:    "test-gpt-4",
		MatchPattern: "gpt-4*",
		StartDate:    time.Now(),
		InputPrice:   0.03,
		OutputPrice:  0.06,
		Unit:         "TOKENS",
		TokenizerId:  "openai",
	}

	fmt.Println("Creating test model...")
	createdModel, err := modelClient.Create(ctx, testModel)
	if err != nil {
		fmt.Printf("Error creating model: %v\n", err)
		return
	}
	fmt.Printf("Created model with ID: %s\n", createdModel.ID)

	// Test listing models
	fmt.Println("Listing models...")
	listParams := models.ListParams{
		Page:  1,
		Limit: 10,
	}
	listResponse, err := modelClient.List(ctx, listParams)
	if err != nil {
		fmt.Printf("Error listing models: %v\n", err)
	} else {
		fmt.Printf("Found %d models\n", len(listResponse.Data))
	}

	// Test getting a specific model
	if createdModel.ID != "" {
		fmt.Printf("Getting model by ID: %s\n", createdModel.ID)
		retrievedModel, err := modelClient.Get(ctx, createdModel.ID)
		if err != nil {
			fmt.Printf("Error getting model: %v\n", err)
		} else {
			fmt.Printf("Retrieved model: %s (match pattern: %s)\n", retrievedModel.ModelName, retrievedModel.MatchPattern)
		}

		// Test deleting the model
		fmt.Printf("Deleting model with ID: %s\n", createdModel.ID)
		err = modelClient.Delete(ctx, createdModel.ID)
		if err != nil {
			fmt.Printf("Error deleting model: %v\n", err)
		} else {
			fmt.Println("Model deleted successfully")
		}
	}

	fmt.Println("Model API tests completed!")
}

func main() {
	langfuseHost := "https://us.cloud.langfuse.com"
	langfusePubKey := os.Getenv("LANGFUSE_PUBLIC_KEY")
	langfuseSecret := os.Getenv("LANGFUSE_SECRET_KEY")

	if langfusePubKey == "" || langfuseSecret == "" {
		fmt.Println("LANGFUSE_PUBLIC_KEY and LANGFUSE_SECRET_KEY environment variables must be set")
		return
	}

	client := langfuse.NewClient(langfuseHost, langfusePubKey, langfuseSecret)
	defer client.Close()

	// Test Traces
	runTraceTests(client)

	// Test Models
	runModelTests(client)
}

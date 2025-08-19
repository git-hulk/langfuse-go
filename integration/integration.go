package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/git-hulk/langfuse-go"
	"github.com/git-hulk/langfuse-go/pkg/annotations"
	"github.com/git-hulk/langfuse-go/pkg/comments"
	"github.com/git-hulk/langfuse-go/pkg/datasets"
	"github.com/git-hulk/langfuse-go/pkg/llmconnections"
	"github.com/git-hulk/langfuse-go/pkg/models"
	"github.com/git-hulk/langfuse-go/pkg/organizations"
	"github.com/git-hulk/langfuse-go/pkg/projects"
	"github.com/git-hulk/langfuse-go/pkg/prompts"
	"github.com/git-hulk/langfuse-go/pkg/scores"
)

// ANSI color codes
const (
	ColorReset = "\033[0m"
	ColorRed   = "\033[31m"
	ColorGreen = "\033[32m"
	ColorBlue  = "\033[34m"
)

// Helper function to print errors in red
func printError(format string, args ...interface{}) {
	fmt.Printf(ColorRed+format+ColorReset, args...)
}

// Helper function to print info messages in blue
func printInfo(format string, args ...interface{}) {
	fmt.Printf(ColorBlue+format+ColorReset, args...)
}

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
		printError("Error creating model: %v\n", err)
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
		printError("Error listing models: %v\n", err)
	} else {
		fmt.Printf("Found %d models\n", len(listResponse.Data))
	}

	// Test getting a specific model
	if createdModel.ID != "" {
		fmt.Printf("Getting model by ID: %s\n", createdModel.ID)
		retrievedModel, err := modelClient.Get(ctx, createdModel.ID)
		if err != nil {
			printError("Error getting model: %v\n", err)
		} else {
			fmt.Printf("Retrieved model: %s (match pattern: %s)\n", retrievedModel.ModelName, retrievedModel.MatchPattern)
		}

		// Test deleting the model
		fmt.Printf("Deleting model with ID: %s\n", createdModel.ID)
		err = modelClient.Delete(ctx, createdModel.ID)
		if err != nil {
			printError("Error deleting model: %v\n", err)
		} else {
			fmt.Println("Model deleted successfully")
		}
	}

	fmt.Println("Model API tests completed!")
}

func runPromptTests(client *langfuse.LangFuse) {
	ctx := context.Background()
	promptClient := client.Prompts()

	fmt.Println("Testing Prompt API...")

	// Test creating a prompt
	testPrompt := &prompts.PromptEntry{
		Name: "test-prompt",
		Type: "chat",
		Prompt: []prompts.ChatMessageWithPlaceHolder{
			{
				Role:    "system",
				Type:    "text",
				Content: "You are a helpful assistant.",
			},
			{
				Role:    "user",
				Type:    "text",
				Content: "Hello {{name}}, how can I help you today?",
			},
		},
		Tags:   []string{"test", "integration"},
		Labels: []string{"v1"},
	}

	fmt.Println("Creating test prompt...")
	createdPrompt, err := promptClient.Create(ctx, testPrompt)
	if err != nil {
		printError("Error creating prompt: %v\n", err)
		return
	}
	fmt.Printf("Created prompt: %s (version: %d)\n", createdPrompt.Name, createdPrompt.Version)

	// Test listing prompts
	fmt.Println("Listing prompts...")
	listParams := prompts.ListParams{
		Page:  1,
		Limit: 10,
	}
	listResponse, err := promptClient.List(ctx, listParams)
	if err != nil {
		printError("Error listing prompts: %v\n", err)
	} else {
		fmt.Printf("Found %d prompts\n", len(listResponse.Data))
	}

	// Test getting a specific prompt
	if createdPrompt.Name != "" {
		fmt.Printf("Getting prompt by name: %s\n", createdPrompt.Name)
		getParams := prompts.GetParams{
			Name:    createdPrompt.Name,
			Version: createdPrompt.Version,
		}
		retrievedPrompt, err := promptClient.Get(ctx, getParams)
		if err != nil {
			printError("Error getting prompt: %v\n", err)
		} else {
			fmt.Printf("Retrieved prompt: %s (type: %s, messages: %d)\n",
				retrievedPrompt.Name, retrievedPrompt.Type, len(retrievedPrompt.Prompt))
		}
	}

	fmt.Println("Prompt API tests completed!")
}

func runScoreTests(client *langfuse.LangFuse) {
	ctx := context.Background()
	scoreClient := client.Scores()

	fmt.Println("Testing Score API...")

	// First create a trace to score against
	trace := client.StartTrace("Score Test Trace")
	trace.Input = map[string]string{"query": "Test query for scoring"}
	trace.Output = map[string]string{"response": "Test response"}
	trace.End()

	// Wait a moment for trace to be processed
	time.Sleep(2 * time.Second)

	// Test creating a score
	testScore := &scores.CreateScoreRequest{
		TraceID:  trace.ID,
		Name:     "test-quality-score",
		DataType: scores.ScoreDataTypeNumeric,
		Value:    0.85,
		Comment:  "Integration test score",
	}

	fmt.Println("Creating test score...")
	createdScore, err := scoreClient.Create(ctx, testScore)
	if err != nil {
		printError("Error creating score: %v\n", err)
		return
	}
	fmt.Printf("Created score with ID: %s\n", createdScore.ID)

	// Test listing scores
	fmt.Println("Listing scores...")
	listParams := scores.ListParams{
		Page:  1,
		Limit: 10,
		Name:  "test-quality-score",
	}
	listResponse, err := scoreClient.List(ctx, listParams)
	if err != nil {
		printError("Error listing scores: %v\n", err)
	} else {
		fmt.Printf("Found %d scores\n", len(listResponse.Data))
	}

	// Test getting a specific score
	if createdScore.ID != "" {
		for i := 0; i < 10; i++ {
			retrievedScore, _ := scoreClient.Get(ctx, createdScore.ID)
			if retrievedScore != nil {
				fmt.Printf("Retrieved score: %s (value: %.2f, comment: %s)\n",
					retrievedScore.ID, retrievedScore.Value, retrievedScore.Comment)
				break
			} else {
				fmt.Printf("Waiting for score to be available... Attempt %d\n", i+1)
			}
			time.Sleep(2 * time.Second)
		}

		// Test deleting the score
		fmt.Printf("Deleting score with ID: %s\n", createdScore.ID)
		err = scoreClient.Delete(ctx, createdScore.ID)
		if err != nil {
			printError("Error deleting score: %v\n", err)
		} else {
			fmt.Println("Score deleted successfully")
		}
	}

	fmt.Println("Score API tests completed!")
}

func runScoreConfigTests(client *langfuse.LangFuse) {
	ctx := context.Background()
	scoreClient := client.Scores()

	fmt.Println("Testing Score Config API...")

	// Test creating a numeric score config
	testNumericConfig := &scores.CreateScoreConfigRequest{
		Name:        "test-numeric-config",
		DataType:    scores.ScoreDataTypeNumeric,
		MinValue:    0.0,
		MaxValue:    1.0,
		Description: "Test numeric score configuration",
	}

	fmt.Println("Creating test numeric score config...")
	createdNumericConfig, err := scoreClient.CreateConfig(ctx, testNumericConfig)
	if err != nil {
		printError("Error creating numeric score config: %v\n", err)
		return
	}
	fmt.Printf("Created numeric score config with ID: %s\n", createdNumericConfig.ID)

	// Test creating a categorical score config
	testCategoricalConfig := &scores.CreateScoreConfigRequest{
		Name:     "test-categorical-config",
		DataType: scores.ScoreDataTypeCategorical,
		Categories: []scores.ConfigCategory{
			{Value: 1.0, Label: "Poor"},
			{Value: 2.0, Label: "Fair"},
			{Value: 3.0, Label: "Good"},
			{Value: 4.0, Label: "Excellent"},
		},
		Description: "Test categorical score configuration",
	}

	fmt.Println("Creating test categorical score config...")
	createdCategoricalConfig, err := scoreClient.CreateConfig(ctx, testCategoricalConfig)
	if err != nil {
		printError("Error creating categorical score config: %v\n", err)
		return
	}
	fmt.Printf("Created categorical score config with ID: %s\n", createdCategoricalConfig.ID)

	// Test creating a boolean score config
	testBooleanConfig := &scores.CreateScoreConfigRequest{
		Name:        "test-boolean-config",
		DataType:    scores.ScoreDataTypeBoolean,
		Description: "Test boolean score configuration",
	}

	fmt.Println("Creating test boolean score config...")
	createdBooleanConfig, err := scoreClient.CreateConfig(ctx, testBooleanConfig)
	if err != nil {
		printError("Error creating boolean score config: %v\n", err)
		return
	}
	fmt.Printf("Created boolean score config with ID: %s\n", createdBooleanConfig.ID)

	// Test listing score configs
	fmt.Println("Listing score configs...")
	configListParams := scores.ConfigListParams{
		Page:  1,
		Limit: 10,
	}
	configListResponse, err := scoreClient.ListConfigs(ctx, configListParams)
	if err != nil {
		printError("Error listing score configs: %v\n", err)
	} else {
		fmt.Printf("Found %d score configs\n", len(configListResponse.Data))
	}

	// Test getting specific score configs
	configIDs := []string{createdNumericConfig.ID, createdCategoricalConfig.ID, createdBooleanConfig.ID}
	configNames := []string{"numeric", "categorical", "boolean"}

	for i, configID := range configIDs {
		if configID != "" {
			fmt.Printf("Getting %s score config by ID: %s\n", configNames[i], configID)
			retrievedConfig, err := scoreClient.GetConfig(ctx, configID)
			if err != nil {
				fmt.Printf("Error getting %s score config: %v\n", configNames[i], err)
			} else {
				fmt.Printf("Retrieved %s score config: %s (type: %s)\n",
					configNames[i], retrievedConfig.Name, retrievedConfig.DataType)
				if retrievedConfig.DataType == scores.ScoreDataTypeCategorical {
					fmt.Printf("  Categories: %d\n", len(retrievedConfig.Categories))
				}
				if retrievedConfig.DataType == scores.ScoreDataTypeNumeric {
					fmt.Printf("  Range: %.1f - %.1f\n", retrievedConfig.MinValue, retrievedConfig.MaxValue)
				}
			}
		}
	}

	fmt.Println("Score Config API tests completed!")
}

func runDatasetTests(client *langfuse.LangFuse) {
	ctx := context.Background()
	datasetClient := client.Datasets()

	fmt.Println("Testing Dataset API...")

	// Test creating a dataset
	testDataset := &datasets.CreateDatasetRequest{
		Name:        "test-integration-dataset",
		Description: "Integration test dataset for Go client",
		Metadata: map[string]interface{}{
			"version": "1.0",
			"source":  "integration-test",
		},
	}

	fmt.Println("Creating test dataset...")
	createdDataset, err := datasetClient.Create(ctx, testDataset)
	if err != nil {
		printError("Error creating dataset: %v\n", err)
		return
	}
	fmt.Printf("Created dataset: %s (ID: %s)\n", createdDataset.Name, createdDataset.ID)

	// Test listing datasets
	fmt.Println("Listing datasets...")
	listParams := datasets.ListParams{
		Page:  1,
		Limit: 10,
	}
	listResponse, err := datasetClient.List(ctx, listParams)
	if err != nil {
		fmt.Printf("Error listing datasets: %v\n", err)
	} else {
		fmt.Printf("Found %d datasets\n", len(listResponse.Data))
	}

	// Test getting a specific dataset
	if createdDataset.Name != "" {
		fmt.Printf("Getting dataset by name: %s\n", createdDataset.Name)
		retrievedDataset, err := datasetClient.Get(ctx, createdDataset.Name)
		if err != nil {
			fmt.Printf("Error getting dataset: %v\n", err)
		} else {
			fmt.Printf("Retrieved dataset: %s (description: %s)\n",
				retrievedDataset.Name, retrievedDataset.Description)
		}
	}

	fmt.Println("Dataset API tests completed!")
}

func runDatasetItemTests(client *langfuse.LangFuse) {
	ctx := context.Background()
	datasetClient := client.Datasets()

	fmt.Println("Testing Dataset Item API...")

	// First, create a dataset for testing items
	testDataset := &datasets.CreateDatasetRequest{
		Name:        "test-item-dataset",
		Description: "Dataset for testing items",
	}

	createdDataset, err := datasetClient.Create(ctx, testDataset)
	if err != nil {
		fmt.Printf("Error creating dataset for items: %v\n", err)
		return
	}

	// Test creating dataset items
	testItems := []*datasets.CreateDatasetItemRequest{
		{
			DatasetName: createdDataset.Name,
			Input: map[string]interface{}{
				"query": "What is the capital of France?",
			},
			ExpectedOutput: map[string]interface{}{
				"answer": "Paris",
			},
			Metadata: map[string]interface{}{
				"category": "geography",
			},
		},
		{
			DatasetName: createdDataset.Name,
			Input: map[string]interface{}{
				"query": "What is 2 + 2?",
			},
			ExpectedOutput: map[string]interface{}{
				"answer": "4",
			},
			Metadata: map[string]interface{}{
				"category": "math",
			},
		},
	}

	var createdItemIDs []string
	for i, item := range testItems {
		fmt.Printf("Creating dataset item %d...\n", i+1)
		createdItem, err := datasetClient.CreateDatasetItem(ctx, item)
		if err != nil {
			fmt.Printf("Error creating dataset item %d: %v\n", i+1, err)
			continue
		}
		fmt.Printf("Created dataset item %d with ID: %s\n", i+1, createdItem.ID)
		createdItemIDs = append(createdItemIDs, createdItem.ID)
	}

	// Test listing dataset items
	fmt.Println("Listing dataset items...")
	itemListParams := datasets.ListDatasetItemParams{
		DatasetName: createdDataset.Name,
		Page:        1,
		Limit:       10,
	}
	itemListResponse, err := datasetClient.ListDatasetItems(ctx, itemListParams)
	if err != nil {
		fmt.Printf("Error listing dataset items: %v\n", err)
	} else {
		fmt.Printf("Found %d dataset items\n", len(itemListResponse.Data))
	}

	// Test getting specific dataset items
	for i, itemID := range createdItemIDs {
		if itemID != "" {
			fmt.Printf("Getting dataset item %d by ID: %s\n", i+1, itemID)
			retrievedItem, err := datasetClient.GetDatasetItem(ctx, itemID)
			if err != nil {
				fmt.Printf("Error getting dataset item %d: %v\n", i+1, err)
			} else {
				fmt.Printf("Retrieved dataset item %d from dataset: %s\n",
					i+1, retrievedItem.DatasetName)
			}
		}
	}

	// Test deleting dataset items
	for i, itemID := range createdItemIDs {
		if itemID != "" {
			fmt.Printf("Deleting dataset item %d with ID: %s\n", i+1, itemID)
			err := datasetClient.DeleteDatasetItem(ctx, itemID)
			if err != nil {
				fmt.Printf("Error deleting dataset item %d: %v\n", i+1, err)
			} else {
				fmt.Printf("Dataset item %d deleted successfully\n", i+1)
			}
		}
	}

	fmt.Println("Dataset Item API tests completed!")
}

func runDatasetRunTests(client *langfuse.LangFuse) {
	ctx := context.Background()
	datasetClient := client.Datasets()

	fmt.Println("Testing Dataset Run API...")

	// First, create a dataset for testing runs
	testDataset := &datasets.CreateDatasetRequest{
		Name:        "test-run-dataset",
		Description: "Dataset for testing runs",
	}

	createdDataset, err := datasetClient.Create(ctx, testDataset)
	if err != nil {
		fmt.Printf("Error creating dataset for runs: %v\n", err)
		return
	}

	// Create some dataset items first
	testItem := &datasets.CreateDatasetItemRequest{
		DatasetName: createdDataset.Name,
		Input: map[string]interface{}{
			"query": "Test query for run",
		},
		ExpectedOutput: map[string]interface{}{
			"answer": "Test expected output",
		},
	}

	_, err = datasetClient.CreateDatasetItem(ctx, testItem)
	if err != nil {
		fmt.Printf("Error creating dataset item for run: %v\n", err)
		return
	}

	// Create a trace to link with the run
	trace := client.StartTrace("Dataset Run Test Trace")
	trace.Input = map[string]string{"query": "Test query for run"}
	trace.Output = map[string]string{"response": "Test response for run"}
	trace.End()

	// Wait a moment for trace to be processed
	time.Sleep(2 * time.Second)

	// Test listing dataset runs (initially empty)
	fmt.Println("Listing dataset runs...")
	runListParams := datasets.ListParams{
		Page:  1,
		Limit: 10,
	}
	runListResponse, err := datasetClient.GetDatasetRuns(ctx, createdDataset.Name, runListParams)
	if err != nil {
		fmt.Printf("Error listing dataset runs: %v\n", err)
	} else {
		fmt.Printf("Found %d dataset runs\n", len(runListResponse.Data))
	}

	fmt.Println("Dataset Run API tests completed!")
}

func runLLMConnectionTests(client *langfuse.LangFuse) {
	ctx := context.Background()
	llmClient := client.LLMConnections()

	fmt.Println("Testing LLM Connection API...")

	// Test creating/updating LLM connections for different adapters
	testConnections := []*llmconnections.UpsertLLMConnectionRequest{
		{
			Provider:          "test-openai-provider",
			Adapter:           llmconnections.AdapterOpenAI,
			SecretKey:         "test-openai-secret-key",
			WithDefaultModels: boolPtr(true),
			CustomModels:      []string{"gpt-4-custom", "gpt-3.5-custom"},
		},
		{
			Provider:          "test-anthropic-provider",
			Adapter:           llmconnections.AdapterAnthropic,
			SecretKey:         "test-anthropic-secret-key",
			WithDefaultModels: boolPtr(true),
			CustomModels:      []string{"claude-3-custom"},
		},
		{
			Provider:     "test-azure-provider",
			Adapter:      llmconnections.AdapterAzure,
			SecretKey:    "test-azure-secret-key",
			BaseURL:      "https://test-azure.openai.azure.com",
			CustomModels: []string{"azure-gpt-4"},
			ExtraHeaders: map[string]string{
				"api-version": "2023-12-01-preview",
			},
		},
		{
			Provider:          "test-bedrock-provider",
			Adapter:           llmconnections.AdapterBedrock,
			SecretKey:         "test-bedrock-secret-key",
			WithDefaultModels: boolPtr(false),
			CustomModels:      []string{"anthropic.claude-3-sonnet-20240229-v1:0"},
		},
		{
			Provider:          "test-vertex-provider",
			Adapter:           llmconnections.AdapterGoogleVertexAI,
			SecretKey:         "test-vertex-secret-key",
			WithDefaultModels: boolPtr(true),
			CustomModels:      []string{"gemini-pro-custom"},
		},
		{
			Provider:          "test-ai-studio-provider",
			Adapter:           llmconnections.AdapterGoogleAIStudio,
			SecretKey:         "test-ai-studio-secret-key",
			WithDefaultModels: boolPtr(true),
		},
	}

	var createdConnections []*llmconnections.LLMConnection
	for i, conn := range testConnections {
		fmt.Printf("Creating/updating LLM connection %d (%s)...\n", i+1, conn.Adapter)
		createdConnection, err := llmClient.Upsert(ctx, conn)
		if err != nil {
			fmt.Printf("Error creating/updating LLM connection %d (%s): %v\n", i+1, conn.Adapter, err)
			continue
		}
		fmt.Printf("Created/updated %s connection with ID: %s\n", createdConnection.Adapter, createdConnection.ID)
		createdConnections = append(createdConnections, createdConnection)
	}

	// Test listing LLM connections
	fmt.Println("Listing LLM connections...")
	listParams := llmconnections.ListParams{
		Page:  1,
		Limit: 10,
	}
	listResponse, err := llmClient.List(ctx, listParams)
	if err != nil {
		fmt.Printf("Error listing LLM connections: %v\n", err)
	} else {
		fmt.Printf("Found %d LLM connections\n", len(listResponse.Data))
		for i, conn := range listResponse.Data {
			fmt.Printf("  %d. %s (%s) - Provider: %s, Models: %v\n",
				i+1, conn.ID, conn.Adapter, conn.Provider, conn.CustomModels)
		}
	}

	// Test updating an existing connection (upsert behavior)
	if len(createdConnections) > 0 {
		fmt.Println("Testing connection update...")
		updateConn := &llmconnections.UpsertLLMConnectionRequest{
			Provider:          createdConnections[0].Provider,
			Adapter:           createdConnections[0].Adapter,
			SecretKey:         "updated-secret-key",
			WithDefaultModels: boolPtr(false),
			CustomModels:      []string{"updated-model-1", "updated-model-2"},
		}

		updatedConnection, err := llmClient.Upsert(ctx, updateConn)
		if err != nil {
			fmt.Printf("Error updating LLM connection: %v\n", err)
		} else {
			fmt.Printf("Updated connection %s with new models: %v\n",
				updatedConnection.ID, updatedConnection.CustomModels)
		}
	}

	fmt.Println("LLM Connection API tests completed!")
}

func runOrganizationTests(client *langfuse.LangFuse) {
	ctx := context.Background()
	fmt.Println("Testing Organization API...")

	// Note: Organization membership APIs require organization-scoped API keys
	// These tests will demonstrate the API usage but may fail with project-scoped keys
	organizationClient := client.Organizations()

	// Test getting organization memberships
	fmt.Println("Listing organization memberships...")
	orgMemberships, err := organizationClient.ListMemberships(ctx)
	if err != nil {
		fmt.Printf("Error listing organization memberships: %v\n", err)
		fmt.Println("Note: Organization membership APIs require organization-scoped API keys")
	} else {
		fmt.Printf("Found %d organization memberships\n", len(orgMemberships.Memberships))
		for i, membership := range orgMemberships.Memberships {
			fmt.Printf("  %d. User: %s (%s) - Role: %s - Email: %s\n",
				i+1, membership.Name, membership.UserID, membership.Role, membership.Email)
		}
	}

	// Test getting project memberships (requires a project ID)
	// We'll use a placeholder project ID for demonstration
	testProjectID := "test-project-id"
	fmt.Printf("Listing project memberships for project: %s...\n", testProjectID)
	projectMemberships, err := organizationClient.ListProjectMemberships(ctx, testProjectID)
	if err != nil {
		fmt.Printf("Error listing project memberships: %v\n", err)
		fmt.Println("Note: This may fail if the project ID doesn't exist or requires organization-scoped API keys")
	} else {
		fmt.Printf("Found %d project memberships\n", len(projectMemberships.Memberships))
		for i, membership := range projectMemberships.Memberships {
			fmt.Printf("  %d. User: %s (%s) - Role: %s - Email: %s\n",
				i+1, membership.Name, membership.UserID, membership.Role, membership.Email)
		}
	}

	// Test updating organization membership (demonstration only)
	fmt.Println("Demonstrating organization membership update request structure...")
	testOrgMembership := &organizations.MembershipRequest{
		UserID: "test-user-id",
		Role:   organizations.MembershipRoleMember,
	}

	fmt.Printf("Sample organization membership update: UserID=%s, Role=%s\n",
		testOrgMembership.UserID, testOrgMembership.Role)

	// Note: We're not actually calling UpdateMembership to avoid
	// unintended membership changes in real organizations
	fmt.Println("Skipping actual membership update to prevent unintended changes")

	// Test updating project membership (demonstration only)
	fmt.Println("Demonstrating project membership update request structure...")
	testProjectMembership := &organizations.MembershipRequest{
		UserID: "test-user-id",
		Role:   organizations.MembershipRoleViewer,
	}

	fmt.Printf("Sample project membership update: ProjectID=%s, UserID=%s, Role=%s\n",
		testProjectID, testProjectMembership.UserID, testProjectMembership.Role)

	// Note: We're not actually calling UpdateProjectMembership to avoid
	// unintended membership changes in real projects
	fmt.Println("Skipping actual membership update to prevent unintended changes")

	// Test role validation
	fmt.Println("Testing membership role constants...")
	roles := []organizations.MembershipRole{
		organizations.MembershipRoleOwner,
		organizations.MembershipRoleAdmin,
		organizations.MembershipRoleMember,
		organizations.MembershipRoleViewer,
	}

	for i, role := range roles {
		fmt.Printf("  %d. Role: %s\n", i+1, role)
	}

	fmt.Println("Organization API tests completed!")
	fmt.Println("Note: Full functionality requires organization-scoped API keys")
}

func runProjectTests(client *langfuse.LangFuse) {
	ctx := context.Background()
	projectClient := client.Projects()

	fmt.Println("Testing Project API...")

	// Test getting current project(s)
	fmt.Println("Getting current project(s)...")
	currentProjects, err := projectClient.Get(ctx)
	if err != nil {
		fmt.Printf("Error getting current projects: %v\n", err)
	} else {
		fmt.Printf("Found %d current project(s)\n", len(currentProjects.Data))
		for i, project := range currentProjects.Data {
			retentionInfo := "no retention set"
			if project.RetentionDays != nil {
				retentionInfo = fmt.Sprintf("%d days", *project.RetentionDays)
			}
			fmt.Printf("  %d. %s (ID: %s) - Retention: %s\n",
				i+1, project.Name, project.ID, retentionInfo)
		}
	}

	// Test creating a new project (requires organization-scoped API key)
	fmt.Println("Demonstrating project creation request structure...")
	testProject := &projects.CreateProjectRequest{
		Name:      "test-integration-project",
		Retention: 30,
		Metadata: map[string]interface{}{
			"environment": "test",
			"purpose":     "integration-testing",
			"version":     "1.0",
		},
	}

	fmt.Printf("Sample project creation: Name=%s, Retention=%d days\n",
		testProject.Name, testProject.Retention)
	fmt.Printf("Metadata: %v\n", testProject.Metadata)

	// Note: We're not actually creating a project to avoid unintended project creation
	fmt.Println("Skipping actual project creation to prevent unintended project creation")
	fmt.Println("Note: Project creation requires organization-scoped API keys")

	// Test updating a project (demonstration only)
	fmt.Println("Demonstrating project update request structure...")
	testUpdate := &projects.UpdateProjectRequest{
		Name:      "updated-test-project",
		Retention: 60,
		Metadata: map[string]interface{}{
			"environment": "production",
			"purpose":     "updated-purpose",
			"version":     "2.0",
		},
	}

	fmt.Printf("Sample project update: Name=%s, Retention=%d days\n",
		testUpdate.Name, testUpdate.Retention)
	fmt.Printf("Updated metadata: %v\n", testUpdate.Metadata)

	fmt.Println("Skipping actual project update to prevent unintended changes")
	fmt.Println("Note: Project updates require organization-scoped API keys")

	// Test API key operations (if we have projects)
	if currentProjects != nil && len(currentProjects.Data) > 0 {
		testProjectID := currentProjects.Data[0].ID

		// Test getting API keys for the project
		fmt.Printf("Getting API keys for project: %s...\n", testProjectID)
		apiKeys, err := projectClient.GetAPIKeys(ctx, testProjectID)
		if err != nil {
			printError("Error getting API keys: %v\n", err)
			fmt.Println("Note: API key management requires organization-scoped API keys")
		} else {
			fmt.Printf("Found %d API key(s) for project\n", len(apiKeys.ApiKeys))
			for i, key := range apiKeys.ApiKeys {
				noteInfo := "no note"
				if key.Note != nil {
					noteInfo = *key.Note
				}
				lastUsedInfo := "never used"
				if key.LastUsedAt != nil {
					lastUsedInfo = key.LastUsedAt.Format("2006-01-02 15:04:05")
				}
				fmt.Printf("  %d. %s (%s) - Note: %s, Last used: %s\n",
					i+1, key.PublicKey, key.DisplaySecretKey, noteInfo, lastUsedInfo)
			}
		}

		// Test creating API key (demonstration only)
		fmt.Println("Demonstrating API key creation request structure...")
		testAPIKeyReq := &projects.CreateAPIKeyRequest{
			Note: stringPtr("Integration test API key"),
		}

		fmt.Printf("Sample API key creation: Note=%s\n", *testAPIKeyReq.Note)

		fmt.Println("Skipping actual API key creation to prevent unintended key creation")
		fmt.Println("Note: API key creation requires organization-scoped API keys")
	}

	// Test project deletion (demonstration only)
	fmt.Println("Demonstrating project deletion (for reference only)...")
	fmt.Println("Project deletion would use projectClient.Delete(ctx, projectID)")
	fmt.Println("Note: Project deletion is processed asynchronously and requires organization-scoped API keys")
	fmt.Println("Skipping actual deletion to prevent unintended project removal")

	fmt.Println("Project API tests completed!")
	fmt.Println("Note: Full project management functionality requires organization-scoped API keys")
}

func runCommentTests(client *langfuse.LangFuse) {
	ctx := context.Background()
	commentClient := client.Comments()

	projectID := os.Getenv("LANGFUSE_PROJECT_ID")
	if projectID == "" {
		fmt.Println("LANGFUSE_PROJECT_ID environment variable is not set. Skipping comment tests.")
		return
	}

	fmt.Println("Testing Comment API...")

	// First, create a trace to comment on
	trace := client.StartTrace("Comment Test Trace")
	trace.Input = map[string]string{"query": "Test query for commenting"}
	trace.Output = map[string]string{"response": "Test response for commenting"}
	trace.End()

	// Wait a moment for trace to be processed
	time.Sleep(2 * time.Second)

	// Test creating comments for different object types
	testComments := []*comments.CreateCommentRequest{
		{
			ProjectID:  projectID,
			ObjectType: comments.ObjectTypeTrace,
			ObjectID:   trace.ID,
			Content:    "This is a test comment on a trace. The trace processed successfully!",
		},
		{
			ProjectID:  projectID,
			ObjectType: comments.ObjectTypeTrace,
			ObjectID:   trace.ID,
			Content:    "Another comment on the same trace with additional feedback.",
		},
	}

	var createdCommentIDs []string
	for i, comment := range testComments {
		fmt.Printf("Creating comment %d on trace...\n", i+1)
		createdComment, err := commentClient.Create(ctx, comment)
		if err != nil {
			printError("Error creating comment %d: %v\n", i+1, err)
			continue
		}
		fmt.Printf("Created comment %d with ID: %s\n", i+1, createdComment.ID)
		fmt.Printf("  Content: %s\n", createdComment.Content)
		createdCommentIDs = append(createdCommentIDs, createdComment.ID)
	}

	// Test listing all comments
	fmt.Println("Listing all comments...")
	listParams := comments.ListParams{
		Page:  1,
		Limit: 10,
	}
	listResponse, err := commentClient.List(ctx, listParams)
	if err != nil {
		fmt.Printf("Error listing comments: %v\n", err)
	} else {
		fmt.Printf("Found %d comments\n", len(listResponse.Data))
		for i, comment := range listResponse.Data {
			fmt.Printf("  %d. %s on %s:%s - %s\n",
				i+1, comment.ID, comment.ObjectType, comment.ObjectID, comment.Content[:50]+"...")
		}
	}

	// Test listing comments filtered by object type and ID
	fmt.Printf("Listing comments for trace: %s...\n", trace.ID)
	filteredParams := comments.ListParams{
		Page:       1,
		Limit:      10,
		ObjectType: comments.ObjectTypeTrace,
		ObjectID:   trace.ID,
	}
	filteredResponse, err := commentClient.List(ctx, filteredParams)
	if err != nil {
		printError("Error listing filtered comments: %v\n", err)
	} else {
		fmt.Printf("Found %d comments for this trace\n", len(filteredResponse.Data))
		for i, comment := range filteredResponse.Data {
			authorInfo := "system"
			if comment.AuthorUserID != nil {
				authorInfo = *comment.AuthorUserID
			}
			fmt.Printf("  %d. Author: %s - %s\n",
				i+1, authorInfo, comment.Content)
		}
	}

	// Test getting specific comments
	for i, commentID := range createdCommentIDs {
		if commentID != "" {
			fmt.Printf("Getting comment %d by ID: %s\n", i+1, commentID)
			retrievedComment, err := commentClient.Get(ctx, commentID)
			if err != nil {
				printError("Error getting comment %d: %v\n", i+1, err)
			} else {
				fmt.Printf("Retrieved comment %d: %s (created: %s)\n",
					i+1, retrievedComment.Content, retrievedComment.CreatedAt.Format("2006-01-02 15:04:05"))
			}
		}
	}

	// Test object type constants
	fmt.Println("Testing comment object type constants...")
	objectTypes := []comments.CommentObjectType{
		comments.ObjectTypeTrace,
		comments.ObjectTypeObservation,
		comments.ObjectTypeSession,
		comments.ObjectTypePrompt,
	}

	for i, objType := range objectTypes {
		fmt.Printf("  %d. Object type: %s\n", i+1, objType)
	}

	// Demonstrate commenting on different object types (structure only)
	fmt.Println("Demonstrating comment structures for different object types...")

	// Example for observation comment
	observationComment := &comments.CreateCommentRequest{
		ProjectID:  projectID,
		ObjectType: comments.ObjectTypeObservation,
		ObjectID:   "example-observation-id",
		Content:    "This observation shows excellent performance metrics.",
	}
	fmt.Printf("Observation comment structure: ObjectType=%s, Content=%s\n",
		observationComment.ObjectType, observationComment.Content)

	// Example for session comment
	sessionComment := &comments.CreateCommentRequest{
		ProjectID:  projectID,
		ObjectType: comments.ObjectTypeSession,
		ObjectID:   "example-session-id",
		Content:    "User session completed successfully with good engagement.",
	}
	fmt.Printf("Session comment structure: ObjectType=%s, Content=%s\n",
		sessionComment.ObjectType, sessionComment.Content)

	// Example for prompt comment
	promptComment := &comments.CreateCommentRequest{
		ProjectID:  projectID,
		ObjectType: comments.ObjectTypePrompt,
		ObjectID:   "example-prompt-id",
		Content:    "This prompt template works well for customer service scenarios.",
	}
	fmt.Printf("Prompt comment structure: ObjectType=%s, Content=%s\n",
		promptComment.ObjectType, promptComment.Content)

	fmt.Println("Comment API tests completed!")
}

func runAnnotationTests(client *langfuse.LangFuse) {
	ctx := context.Background()

	// Create annotation clients directly since they're not exposed through the main client
	langfuseHost := "https://us.cloud.langfuse.com"
	langfusePubKey := os.Getenv("LANGFUSE_PUBLIC_KEY")
	langfuseSecret := os.Getenv("LANGFUSE_SECRET_KEY")

	restyCli := resty.New().
		SetBaseURL(langfuseHost+"/api/public").
		SetBasicAuth(langfusePubKey, langfuseSecret)

	queueClient := annotations.NewQueueClient(restyCli)
	itemClient := annotations.NewItemClient(restyCli)

	fmt.Println("Testing Annotation API...")

	// First, we need score configs for the annotation queue
	scoreClient := client.Scores()

	// Create a simple score config for annotation queue
	testScoreConfig := &scores.CreateScoreConfigRequest{
		Name:        "test-annotation-score",
		DataType:    scores.ScoreDataTypeNumeric,
		MinValue:    1.0,
		MaxValue:    5.0,
		Description: "Test score config for annotation queue",
	}

	fmt.Println("Creating score config for annotation queue...")
	createdScoreConfig, err := scoreClient.CreateConfig(ctx, testScoreConfig)
	if err != nil {
		printError("Error creating score config: %v\n", err)
		printInfo("Note: Score config creation may require specific permissions\n")
		return
	}
	fmt.Printf("Created score config: %s (ID: %s)\n", createdScoreConfig.Name, createdScoreConfig.ID)

	// No quota to test create queue API
	testQueueID := "cmei1f6e00e7uad07dirm3e4n"

	// Test listing annotation queues
	fmt.Println("Listing annotation queues...")
	queueListParams := annotations.QueueListParams{
		Page:  1,
		Limit: 10,
	}
	queueListResponse, err := queueClient.List(ctx, queueListParams)
	if err != nil {
		printError("Error listing annotation queues: %v\n", err)
	} else {
		fmt.Printf("Found %d annotation queues\n", len(queueListResponse.Data))
		for i, queue := range queueListResponse.Data {
			fmt.Printf("  %d. %s (ID: %s) - Score configs: %v\n",
				i+1, queue.Name, queue.ID, queue.ScoreConfigIDs)
		}
	}

	// Test getting specific annotation queue
	fmt.Printf("Getting annotation queue by ID: %s\n", testQueueID)
	retrievedQueue, err := queueClient.Get(ctx, testQueueID)
	if err != nil {
		printError("Error getting annotation queue: %v\n", err)
	} else {
		fmt.Printf("Retrieved queue: %s (description: %s)\n",
			retrievedQueue.Name, retrievedQueue.Description)
	}

	// Create a trace to add to the annotation queue
	trace := client.StartTrace("Annotation Test Trace")
	trace.Input = map[string]string{"query": "Test query for annotation"}
	trace.Output = map[string]string{"response": "Test response for annotation"}
	trace.End()

	// Wait a moment for trace to be processed
	time.Sleep(2 * time.Second)

	// Test creating annotation queue items
	testItems := []*annotations.CreateItemRequest{
		{
			ObjectID:   trace.ID,
			ObjectType: annotations.ObjectTypeTrace,
			Status:     annotations.StatusPending,
		},
	}

	var createdItemIDs []string
	for i, item := range testItems {
		fmt.Printf("Creating annotation queue item %d...\n", i+1)
		createdItem, err := itemClient.Create(ctx, testQueueID, item)
		if err != nil {
			printError("Error creating annotation queue item %d: %v\n", i+1, err)
			continue
		}
		fmt.Printf("Created item %d with ID: %s (status: %s)\n",
			i+1, createdItem.ID, createdItem.Status)
		createdItemIDs = append(createdItemIDs, createdItem.ID)
	}

	// Test listing annotation queue items
	fmt.Printf("Listing items in annotation queue: %s...\n", testQueueID)
	itemListParams := annotations.ItemListParams{
		Page:  1,
		Limit: 10,
	}
	itemListResponse, err := itemClient.List(ctx, testQueueID, itemListParams)
	if err != nil {
		printError("Error listing annotation queue items: %v\n", err)
	} else {
		fmt.Printf("Found %d items in annotation queue\n", len(itemListResponse.Data))
		for i, item := range itemListResponse.Data {
			fmt.Printf("  %d. Item %s - Object: %s:%s (status: %s)\n",
				i+1, item.ID, item.ObjectType, item.ObjectID, item.Status)
		}
	}

	// Test getting specific annotation queue items
	for i, itemID := range createdItemIDs {
		if itemID != "" {
			fmt.Printf("Getting annotation queue item %d by ID: %s\n", i+1, itemID)
			retrievedItem, err := itemClient.Get(ctx, testQueueID, itemID)
			if err != nil {
				printError("Error getting annotation queue item %d: %v\n", i+1, err)
			} else {
				fmt.Printf("Retrieved item %d: %s (created: %s)\n",
					i+1, retrievedItem.ID, retrievedItem.CreatedAt.Format("2006-01-02 15:04:05"))
			}
		}
	}

	// Test updating annotation queue item status
	if len(createdItemIDs) > 0 {
		fmt.Printf("Updating annotation queue item status to COMPLETED...\n")
		updateRequest := &annotations.UpdateItemRequest{
			Status: annotations.StatusCompleted,
		}

		updatedItem, err := itemClient.Update(ctx, testQueueID, createdItemIDs[0], updateRequest)
		if err != nil {
			printError("Error updating annotation queue item: %v\n", err)
		} else {
			fmt.Printf("Updated item status: %s (completed at: %s)\n",
				updatedItem.Status, updatedItem.CompletedAt.Format("2006-01-02 15:04:05"))
		}
	}

	// Test assignment operations (demonstration only)
	fmt.Println("Demonstrating annotation queue assignment operations...")
	testAssignment := &annotations.AssignmentRequest{
		UserID: "test-user-id",
	}

	fmt.Printf("Sample assignment request: UserID=%s\n", testAssignment.UserID)

	// Note: We're not actually creating assignments to avoid unintended user assignments
	fmt.Println("Skipping actual assignment creation to prevent unintended user assignments")

	// Test object type and status constants
	fmt.Println("Testing annotation constants...")
	objectTypes := []annotations.QueueObjectType{
		annotations.ObjectTypeTrace,
		annotations.ObjectTypeObservation,
	}

	statuses := []annotations.QueueStatus{
		annotations.StatusPending,
		annotations.StatusCompleted,
	}

	fmt.Println("Object types:")
	for i, objType := range objectTypes {
		fmt.Printf("  %d. %s\n", i+1, objType)
	}

	fmt.Println("Queue statuses:")
	for i, status := range statuses {
		fmt.Printf("  %d. %s\n", i+1, status)
	}

	// Cleanup: Delete created items
	for i, itemID := range createdItemIDs {
		if itemID != "" {
			fmt.Printf("Deleting annotation queue item %d with ID: %s\n", i+1, itemID)
			deleteResponse, err := itemClient.Delete(ctx, testQueueID, itemID)
			if err != nil {
				printError("Error deleting annotation queue item %d: %v\n", i+1, err)
			} else {
				fmt.Printf("Deletion response: %s\n", deleteResponse.Message)
			}
		}
	}

	fmt.Println("Annotation API tests completed!")
	fmt.Printf("Note: Created annotation queue %s may need manual cleanup\n", testQueueID)
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
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

	printInfo("================== TRACE TESTS BEGIN ==================\n")
	runTraceTests(client)
	printInfo("================== TRACE TESTS END ==================\n")

	printInfo("================== MODEL TESTS BEGIN ==================\n")
	runModelTests(client)
	printInfo("================== MODEL TESTS END ==================\n")

	printInfo("================== PROMPT TESTS BEGIN ==================\n")
	runPromptTests(client)
	printInfo("================== PROMPT TESTS END ==================\n")

	printInfo("================== SCORE TESTS BEGIN ==================\n")
	runScoreTests(client)
	printInfo("================== SCORE TESTS END ==================\n")

	printInfo("================== SCORE CONFIG TESTS BEGIN ==================\n")
	runScoreConfigTests(client)
	printInfo("================== SCORE CONFIG TESTS END ==================\n")

	printInfo("================== DATASET TESTS BEGIN ==================\n")
	runDatasetTests(client)
	printInfo("================== DATASET TESTS END ==================\n")

	printInfo("================== DATASET ITEM TESTS BEGIN ==================\n")
	runDatasetItemTests(client)
	printInfo("================== DATASET ITEM TESTS END ==================\n")

	printInfo("================== DATASET RUN TESTS BEGIN ==================\n")
	runDatasetRunTests(client)
	printInfo("================== DATASET RUN TESTS END ==================\n")

	printInfo("================== LLM CONNECTION TESTS BEGIN ==================\n")
	runLLMConnectionTests(client)
	printInfo("================== LLM CONNECTION TESTS END ==================\n")

	printInfo("================== ORGANIZATION TESTS BEGIN ==================\n")
	runOrganizationTests(client)
	printInfo("================== ORGANIZATION TESTS END ==================\n")

	printInfo("================== PROJECT TESTS BEGIN ==================\n")
	runProjectTests(client)
	printInfo("================== PROJECT TESTS END ==================\n")

	printInfo("================== COMMENT TESTS BEGIN ==================\n")
	runCommentTests(client)
	printInfo("================== COMMENT TESTS END ==================\n")

	printInfo("================== ANNOTATION TESTS BEGIN ==================\n")
	runAnnotationTests(client)
	printInfo("================== ANNOTATION TESTS END ==================\n")
}

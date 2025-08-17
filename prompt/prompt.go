package prompt

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"strconv"
	"time"
)

type Prompt struct {
	Name              string    `json:"name"`
	Version           int       `json:"version"`
	Prompt            string    `json:"prompt"`
	IsActive          bool      `json:"isActive"`
	LangfuseCreatedAt time.Time `json:"langfuseCreatedAt"`
	LangfuseUpdatedAt time.Time `json:"langfuseUpdatedAt"`
}

type PromptClient struct {
	restyCli *resty.Client
}

func NewPromptClient(cli *resty.Client) *PromptClient {
	return &PromptClient{restyCli: cli}
}

func (c *PromptClient) Get(name string, version int, label string) (*Prompt, error) {
	var prompt Prompt
	var err error

	req := c.restyCli.R().SetResult(&prompt)

	if version > 0 {
		req.SetQueryParam("version", strconv.Itoa(version))
	}
	if label != "" {
		req.SetQueryParam("label", label)
	}
	if name != "" {
		req.SetQueryParam("name", name)
	} else {
		return nil, fmt.Errorf("prompt name is required")
	}

	resp, err := req.Get("/v2/prompts")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to get prompt: %s", resp.String())
	}
	return &prompt, nil
}

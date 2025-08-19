package comments

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/git-hulk/langfuse-go/pkg/common"

	"github.com/go-resty/resty/v2"
)

// CommentObjectType represents the type of object that can have comments.
type CommentObjectType string

const (
	ObjectTypeTrace       CommentObjectType = "trace"
	ObjectTypeObservation CommentObjectType = "observation"
	ObjectTypeSession     CommentObjectType = "session"
	ObjectTypePrompt      CommentObjectType = "prompt"
)

// CommentEntry represents a Langfuse comment.
type CommentEntry struct {
	ID           string            `json:"id,omitempty"`
	ProjectID    string            `json:"projectId,omitempty"`
	CreatedAt    time.Time         `json:"createdAt,omitempty"`
	UpdatedAt    time.Time         `json:"updatedAt,omitempty"`
	ObjectType   CommentObjectType `json:"objectType"`
	ObjectID     string            `json:"objectId"`
	Content      string            `json:"content"`
	AuthorUserID *string           `json:"authorUserId,omitempty"`
}

func (c *CommentEntry) validate() error {
	if c.ObjectType == "" {
		return errors.New("'objectType' is required")
	}
	if c.ObjectID == "" {
		return errors.New("'objectId' is required")
	}
	if c.Content == "" {
		return errors.New("'content' is required")
	}
	return nil
}

// CreateCommentRequest represents the request to create a comment.
type CreateCommentRequest struct {
	ProjectID    string            `json:"projectId,omitempty"`
	ObjectType   CommentObjectType `json:"objectType"`
	ObjectID     string            `json:"objectId"`
	Content      string            `json:"content"`
	AuthorUserID *string           `json:"authorUserId,omitempty"`
}

func (c *CreateCommentRequest) validate() error {
	if c.ProjectID == "" {
		return errors.New("'projectID' is required")
	}
	if c.ObjectType == "" {
		return errors.New("'objectType' is required")
	}
	if c.ObjectID == "" {
		return errors.New("'objectID' is required")
	}
	if c.Content == "" {
		return errors.New("'content' is required")
	}
	return nil
}

// ListParams defines the query parameters for listing comments.
type ListParams struct {
	Page       int
	Limit      int
	ObjectType CommentObjectType
	ObjectID   string
}

// ToQueryString converts the ListParams to a URL query string.
func (query *ListParams) ToQueryString() string {
	parts := make([]string, 0)
	if query.Page != 0 {
		parts = append(parts, "page="+strconv.Itoa(query.Page))
	}
	if query.Limit != 0 {
		parts = append(parts, "limit="+strconv.Itoa(query.Limit))
	}
	if query.ObjectType != "" {
		parts = append(parts, "objectType="+string(query.ObjectType))
	}
	if query.ObjectID != "" {
		parts = append(parts, "objectId="+query.ObjectID)
	}
	return strings.Join(parts, "&")
}

// ListComments represents the response from listing comments.
type ListComments struct {
	Metadata common.ListMetadata `json:"meta"`
	Data     []CommentEntry      `json:"data"`
}

// Client represents the comments API client.
type Client struct {
	restyCli *resty.Client
}

// NewClient creates a new comments API client.
func NewClient(cli *resty.Client) *Client {
	return &Client{restyCli: cli}
}

// Get retrieves a specific comment by ID.
func (c *Client) Get(ctx context.Context, id string) (*CommentEntry, error) {
	if id == "" {
		return nil, errors.New("'id' is required")
	}

	var comment CommentEntry
	req := c.restyCli.R().
		SetContext(ctx).
		SetResult(&comment).
		SetPathParam("id", id)

	rsp, err := req.Get("/comments/{id}")
	if err != nil {
		return nil, err
	}
	if rsp.IsError() {
		return nil, fmt.Errorf("get comment failed with status code %d", rsp.StatusCode())
	}
	return &comment, nil
}

// List retrieves a list of comments based on the provided parameters.
func (c *Client) List(ctx context.Context, params ListParams) (*ListComments, error) {
	var listResponse ListComments
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&listResponse).
		SetQueryString(params.ToQueryString()).
		Get("/comments")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("list comments failed with status code %d", rsp.StatusCode())
	}
	return &listResponse, nil
}

// Create creates a new comment.
func (c *Client) Create(ctx context.Context, createComment *CreateCommentRequest) (*CommentEntry, error) {
	if err := createComment.validate(); err != nil {
		return nil, err
	}

	var createdComment CommentEntry
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(createComment).
		SetResult(&createdComment).
		Post("/comments")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("failed to create comment, got status code: %d",
			rsp.StatusCode())
	}
	return &createdComment, nil
}

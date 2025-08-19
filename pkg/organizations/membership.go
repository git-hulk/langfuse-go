package organizations

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// MembershipRole represents the role of a membership.
type MembershipRole string

const (
	MembershipRoleOwner  MembershipRole = "OWNER"
	MembershipRoleAdmin  MembershipRole = "ADMIN"
	MembershipRoleMember MembershipRole = "MEMBER"
	MembershipRoleViewer MembershipRole = "VIEWER"
)

// MembershipRequest represents a request to create or update a membership.
type MembershipRequest struct {
	UserID string         `json:"userId"`
	Role   MembershipRole `json:"role"`
}

func (m *MembershipRequest) validate() error {
	if m.UserID == "" {
		return errors.New("'userId' is required")
	}
	if m.Role == "" {
		return errors.New("'role' is required")
	}
	return nil
}

// MembershipResponse represents a membership response.
type MembershipResponse struct {
	UserID string         `json:"userId"`
	Role   MembershipRole `json:"role"`
	Email  string         `json:"email"`
	Name   string         `json:"name"`
}

// MembershipsResponse represents the response from listing memberships.
type MembershipsResponse struct {
	Memberships []MembershipResponse `json:"memberships"`
}

// Client represents the organization memberships API client.
type Client struct {
	restyCli *resty.Client
}

// NewClient creates a new organization memberships API client.
func NewClient(cli *resty.Client) *Client {
	return &Client{restyCli: cli}
}

// ListMemberships retrieves all memberships for the organization associated with the API key.
// Requires organization-scoped API key.
func (c *Client) ListMemberships(ctx context.Context) (*MembershipsResponse, error) {
	var memberships MembershipsResponse
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&memberships).
		Get("/organizations/memberships")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("get organization memberships failed with status code %d", rsp.StatusCode())
	}
	return &memberships, nil
}

// UpdateMembership creates or updates a membership for the organization associated with the API key.
// Requires organization-scoped API key.
func (c *Client) UpdateMembership(ctx context.Context, membership *MembershipRequest) (*MembershipResponse, error) {
	if err := membership.validate(); err != nil {
		return nil, err
	}

	var updatedMembership MembershipResponse
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(membership).
		SetResult(&updatedMembership).
		Put("/organizations/memberships")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("failed to update organization membership: %s, got status code: %d",
			rsp.String(), rsp.StatusCode())
	}
	return &updatedMembership, nil
}

// ListProjectMemberships retrieves all memberships for a specific project.
// Requires organization-scoped API key.
func (c *Client) ListProjectMemberships(ctx context.Context, projectId string) (*MembershipsResponse, error) {
	if projectId == "" {
		return nil, errors.New("'projectId' is required")
	}

	var memberships MembershipsResponse
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&memberships).
		SetPathParam("projectId", projectId).
		Get("/projects/{projectId}/memberships")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("get project memberships failed with status code %d", rsp.StatusCode())
	}
	return &memberships, nil
}

// UpdateProjectMembership creates or updates a membership for a specific project.
// The user must already be a member of the organization.
// Requires organization-scoped API key.
func (c *Client) UpdateProjectMembership(ctx context.Context, projectId string, membership *MembershipRequest) (*MembershipResponse, error) {
	if projectId == "" {
		return nil, errors.New("'projectId' is required")
	}
	if err := membership.validate(); err != nil {
		return nil, err
	}

	var updatedMembership MembershipResponse
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(membership).
		SetResult(&updatedMembership).
		SetPathParam("projectId", projectId).
		Put("/projects/{projectId}/memberships")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("failed to update project membershipx, got status code: %d", rsp.StatusCode())
	}
	return &updatedMembership, nil
}

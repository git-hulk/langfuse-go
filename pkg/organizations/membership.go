package organizations

import (
	"context"
	"errors"
	"fmt"
	"net/http"

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

// MembershipClient represents the organization memberships API client.
type MembershipClient struct {
	restyCli *resty.Client
}

// NewMembershipClient creates a new organization memberships API client.
func NewMembershipClient(cli *resty.Client) *MembershipClient {
	return &MembershipClient{restyCli: cli}
}

// GetOrganizationMemberships retrieves all memberships for the organization associated with the API key.
// Requires organization-scoped API key.
func (c *MembershipClient) GetOrganizationMemberships(ctx context.Context) (*MembershipsResponse, error) {
	var memberships MembershipsResponse
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&memberships).
		Get("/organizations/memberships")
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get organization memberships failed with status code %d", rsp.StatusCode())
	}
	return &memberships, nil
}

// UpdateOrganizationMembership creates or updates a membership for the organization associated with the API key.
// Requires organization-scoped API key.
func (c *MembershipClient) UpdateOrganizationMembership(ctx context.Context, membership *MembershipRequest) (*MembershipResponse, error) {
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

	if rsp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to update organization membership: %s, got status code: %d",
			rsp.String(), rsp.StatusCode())
	}
	return &updatedMembership, nil
}

// GetProjectMemberships retrieves all memberships for a specific project.
// Requires organization-scoped API key.
func (c *MembershipClient) GetProjectMemberships(ctx context.Context, projectId string) (*MembershipsResponse, error) {
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

	if rsp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get project memberships failed with status code %d", rsp.StatusCode())
	}
	return &memberships, nil
}

// UpdateProjectMembership creates or updates a membership for a specific project.
// The user must already be a member of the organization.
// Requires organization-scoped API key.
func (c *MembershipClient) UpdateProjectMembership(ctx context.Context, projectId string, membership *MembershipRequest) (*MembershipResponse, error) {
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

	if rsp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to update project membership: %s, got status code: %d",
			rsp.String(), rsp.StatusCode())
	}
	return &updatedMembership, nil
}

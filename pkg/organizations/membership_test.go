package organizations

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestMembershipRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		req     MembershipRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: MembershipRequest{
				UserID: "user123",
				Role:   MembershipRoleAdmin,
			},
			wantErr: false,
		},
		{
			name: "missing userId",
			req: MembershipRequest{
				Role: MembershipRoleAdmin,
			},
			wantErr: true,
			errMsg:  "'userId' is required",
		},
		{
			name: "missing role",
			req: MembershipRequest{
				UserID: "user123",
			},
			wantErr: true,
			errMsg:  "'role' is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMembershipClient_GetOrganizationMemberships(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/organizations/memberships", r.URL.Path)
		require.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		response := MembershipsResponse{
			Memberships: []MembershipResponse{
				{
					UserID: "user1",
					Role:   MembershipRoleAdmin,
					Email:  "user1@example.com",
					Name:   "User One",
				},
				{
					UserID: "user2",
					Role:   MembershipRoleMember,
					Email:  "user2@example.com",
					Name:   "User Two",
				},
			},
		}
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	membershipClient := NewMembershipClient(client)

	ctx := context.Background()
	memberships, err := membershipClient.GetOrganizationMemberships(ctx)

	require.NoError(t, err)
	require.NotNil(t, memberships)
	require.Len(t, memberships.Memberships, 2)
	require.Equal(t, "user1", memberships.Memberships[0].UserID)
	require.Equal(t, MembershipRoleAdmin, memberships.Memberships[0].Role)
	require.Equal(t, "user1@example.com", memberships.Memberships[0].Email)
	require.Equal(t, "User One", memberships.Memberships[0].Name)
}

func TestMembershipClient_UpdateOrganizationMembership(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/organizations/memberships", r.URL.Path)
		require.Equal(t, "PUT", r.Method)

		var req MembershipRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.Equal(t, "newuser", req.UserID)
		require.Equal(t, MembershipRoleAdmin, req.Role)

		w.Header().Set("Content-Type", "application/json")
		response := MembershipResponse{
			UserID: req.UserID,
			Role:   req.Role,
			Email:  "test@example.com",
			Name:   "Test User",
		}
		err = json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	membershipClient := NewMembershipClient(client)

	ctx := context.Background()
	req := &MembershipRequest{
		UserID: "newuser",
		Role:   MembershipRoleAdmin,
	}

	membership, err := membershipClient.UpdateOrganizationMembership(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, membership)
	require.Equal(t, "newuser", membership.UserID)
	require.Equal(t, MembershipRoleAdmin, membership.Role)
	require.Equal(t, "test@example.com", membership.Email)
	require.Equal(t, "Test User", membership.Name)
}

func TestMembershipClient_UpdateOrganizationMembership_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should not reach here due to validation error
		t.Fatal("Should not reach server due to validation error")
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	membershipClient := NewMembershipClient(client)

	ctx := context.Background()
	req := &MembershipRequest{
		Role: MembershipRoleAdmin,
		// Missing UserID
	}

	membership, err := membershipClient.UpdateOrganizationMembership(ctx, req)

	require.Error(t, err)
	require.Nil(t, membership)
	require.Contains(t, err.Error(), "'userId' is required")
}

func TestMembershipClient_GetProjectMemberships(t *testing.T) {
	projectID := "project1"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/projects/"+projectID+"/memberships", r.URL.Path)
		require.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		response := MembershipsResponse{
			Memberships: []MembershipResponse{
				{
					UserID: "user3",
					Role:   MembershipRoleViewer,
					Email:  "user3@example.com",
					Name:   "User Three",
				},
			},
		}
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	membershipClient := NewMembershipClient(client)

	ctx := context.Background()
	memberships, err := membershipClient.GetProjectMemberships(ctx, projectID)

	require.NoError(t, err)
	require.NotNil(t, memberships)
	require.Len(t, memberships.Memberships, 1)
	require.Equal(t, "user3", memberships.Memberships[0].UserID)
	require.Equal(t, MembershipRoleViewer, memberships.Memberships[0].Role)
}

func TestMembershipClient_GetProjectMemberships_EmptyProjectId(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should not reach here due to validation error
		t.Fatal("Should not reach server due to validation error")
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	membershipClient := NewMembershipClient(client)

	ctx := context.Background()
	memberships, err := membershipClient.GetProjectMemberships(ctx, "")

	require.Error(t, err)
	require.Nil(t, memberships)
	require.Contains(t, err.Error(), "'projectId' is required")
}

func TestMembershipClient_UpdateProjectMembership(t *testing.T) {
	projectID := "project1"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/projects/"+projectID+"/memberships", r.URL.Path)
		require.Equal(t, "PUT", r.Method)

		var req MembershipRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.Equal(t, "projectuser", req.UserID)
		require.Equal(t, MembershipRoleMember, req.Role)

		w.Header().Set("Content-Type", "application/json")
		response := MembershipResponse{
			UserID: req.UserID,
			Role:   req.Role,
			Email:  "projectuser@example.com",
			Name:   "Project User",
		}
		err = json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	membershipClient := NewMembershipClient(client)

	ctx := context.Background()
	req := &MembershipRequest{
		UserID: "projectuser",
		Role:   MembershipRoleMember,
	}

	membership, err := membershipClient.UpdateProjectMembership(ctx, projectID, req)

	require.NoError(t, err)
	require.NotNil(t, membership)
	require.Equal(t, "projectuser", membership.UserID)
	require.Equal(t, MembershipRoleMember, membership.Role)
	require.Equal(t, "projectuser@example.com", membership.Email)
	require.Equal(t, "Project User", membership.Name)
}

func TestMembershipClient_UpdateProjectMembership_EmptyProjectId(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should not reach here due to validation error
		t.Fatal("Should not reach server due to validation error")
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	membershipClient := NewMembershipClient(client)

	ctx := context.Background()
	req := &MembershipRequest{
		UserID: "projectuser",
		Role:   MembershipRoleMember,
	}

	membership, err := membershipClient.UpdateProjectMembership(ctx, "", req)

	require.Error(t, err)
	require.Nil(t, membership)
	require.Contains(t, err.Error(), "'projectId' is required")
}

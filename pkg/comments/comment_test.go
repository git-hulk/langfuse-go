package comments

import (
	"context"
	"testing"
)

func TestCommentEntry_validate(t *testing.T) {
	tests := []struct {
		name    string
		comment CommentEntry
		wantErr bool
	}{
		{
			name: "valid comment",
			comment: CommentEntry{
				ObjectType: ObjectTypeTrace,
				ObjectID:   "trace-123",
				Content:    "This is a test comment",
			},
			wantErr: false,
		},
		{
			name: "missing object type",
			comment: CommentEntry{
				ObjectID: "trace-123",
				Content:  "This is a test comment",
			},
			wantErr: true,
		},
		{
			name: "missing object id",
			comment: CommentEntry{
				ObjectType: ObjectTypeTrace,
				Content:    "This is a test comment",
			},
			wantErr: true,
		},
		{
			name: "missing content",
			comment: CommentEntry{
				ObjectType: ObjectTypeTrace,
				ObjectID:   "trace-123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.comment.validate(); (err != nil) != tt.wantErr {
				t.Errorf("CommentEntry.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateCommentRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		request CreateCommentRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: CreateCommentRequest{
				ProjectID:  "project-123",
				ObjectType: ObjectTypeTrace,
				ObjectID:   "trace-123",
				Content:    "This is a test comment",
			},
			wantErr: false,
		},
		{
			name: "missing object type",
			request: CreateCommentRequest{
				ProjectID: "project-123",
				ObjectID:  "trace-123",
				Content:   "This is a test comment",
			},
			wantErr: true,
		},
		{
			name: "missing object id",
			request: CreateCommentRequest{
				ProjectID:  "project-123",
				ObjectType: ObjectTypeTrace,
				Content:    "This is a test comment",
			},
			wantErr: true,
		},
		{
			name: "missing content",
			request: CreateCommentRequest{
				ProjectID:  "project-123",
				ObjectType: ObjectTypeTrace,
				ObjectID:   "trace-123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.request.validate(); (err != nil) != tt.wantErr {
				t.Errorf("CreateCommentRequest.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListParams_ToQueryString(t *testing.T) {
	tests := []struct {
		name   string
		params ListParams
		want   string
	}{
		{
			name:   "empty params",
			params: ListParams{},
			want:   "",
		},
		{
			name: "all params",
			params: ListParams{
				Page:       1,
				Limit:      10,
				ObjectType: ObjectTypeTrace,
				ObjectID:   "trace-123",
			},
			want: "page=1&limit=10&objectType=TRACE&objectId=trace-123",
		},
		{
			name: "partial params",
			params: ListParams{
				Page:  2,
				Limit: 20,
			},
			want: "page=2&limit=20",
		},
		{
			name: "object type only",
			params: ListParams{
				ObjectType: ObjectTypeObservation,
			},
			want: "objectType=OBSERVATION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.params.ToQueryString(); got != tt.want {
				t.Errorf("ListParams.ToQueryString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommentObjectType(t *testing.T) {
	tests := []struct {
		name       string
		objectType CommentObjectType
		want       string
	}{
		{
			name:       "trace type",
			objectType: ObjectTypeTrace,
			want:       "TRACE",
		},
		{
			name:       "observation type",
			objectType: ObjectTypeObservation,
			want:       "OBSERVATION",
		},
		{
			name:       "session type",
			objectType: ObjectTypeSession,
			want:       "SESSION",
		},
		{
			name:       "prompt type",
			objectType: ObjectTypePrompt,
			want:       "PROMPT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.objectType); got != tt.want {
				t.Errorf("CommentObjectType = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientMethods(t *testing.T) {
	client := NewClient(nil)
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	ctx := context.Background()

	t.Run("GetDatasetItem requires ID", func(t *testing.T) {
		_, err := client.Get(ctx, "")
		if err == nil {
			t.Error("GetDatasetItem() should return error when ID is empty")
		}
	})

	t.Run("CreateDatasetItem validates request", func(t *testing.T) {
		_, err := client.Create(ctx, &CreateCommentRequest{})
		if err == nil {
			t.Error("CreateDatasetItem() should return error for invalid request")
		}

		validRequest := &CreateCommentRequest{
			ProjectID:  "test-project",
			ObjectType: ObjectTypeTrace,
			ObjectID:   "trace-123",
			Content:    "Test comment",
		}
		if err := validRequest.validate(); err != nil {
			t.Errorf("Valid request should not have validation error: %v", err)
		}
	})
}

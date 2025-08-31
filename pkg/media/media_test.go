package media

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestGetMediaUploadURLRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		request GetMediaUploadURLRequest
		wantErr bool
		errMsg  string
	}{
		{
			"valid request",
			GetMediaUploadURLRequest{
				TraceID:       "trace-123",
				ContentType:   ContentTypeImagePNG,
				ContentLength: 1024,
				SHA256Hash:    "abcd1234",
				Field:         "input",
			},
			false,
			"",
		},
		{
			"missing trace id",
			GetMediaUploadURLRequest{
				ContentType:   ContentTypeImagePNG,
				ContentLength: 1024,
				SHA256Hash:    "abcd1234",
				Field:         "input",
			},
			true,
			"'traceId' is required",
		},
		{
			"missing content type",
			GetMediaUploadURLRequest{
				TraceID:       "trace-123",
				ContentLength: 1024,
				SHA256Hash:    "abcd1234",
				Field:         "input",
			},
			true,
			"'contentType' is required",
		},
		{
			"invalid content length",
			GetMediaUploadURLRequest{
				TraceID:       "trace-123",
				ContentType:   ContentTypeImagePNG,
				ContentLength: 0,
				SHA256Hash:    "abcd1234",
				Field:         "input",
			},
			true,
			"'contentLength' must be greater than 0",
		},
		{
			"missing sha256 hash",
			GetMediaUploadURLRequest{
				TraceID:       "trace-123",
				ContentType:   ContentTypeImagePNG,
				ContentLength: 1024,
				Field:         "input",
			},
			true,
			"'sha256Hash' is required",
		},
		{
			"missing field",
			GetMediaUploadURLRequest{
				TraceID:       "trace-123",
				ContentType:   ContentTypeImagePNG,
				ContentLength: 1024,
				SHA256Hash:    "abcd1234",
			},
			true,
			"'field' is required",
		},
		{
			"invalid field",
			GetMediaUploadURLRequest{
				TraceID:       "trace-123",
				ContentType:   ContentTypeImagePNG,
				ContentLength: 1024,
				SHA256Hash:    "abcd1234",
				Field:         "invalid",
			},
			true,
			"'field' must be one of: input, output, metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPatchMediaRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		request PatchMediaRequest
		wantErr bool
		errMsg  string
	}{
		{
			"valid request",
			PatchMediaRequest{
				UploadedAt:       time.Now(),
				UploadHTTPStatus: 200,
			},
			false,
			"",
		},
		{
			"missing uploaded at",
			PatchMediaRequest{
				UploadHTTPStatus: 200,
			},
			true,
			"'uploadedAt' is required",
		},
		{
			"missing upload http status",
			PatchMediaRequest{
				UploadedAt: time.Now(),
			},
			true,
			"'uploadHttpStatus' is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_GetUploadURL(t *testing.T) {
	mockUploadURL := "https://example.com/upload"
	mockMediaID := "media-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "/media", r.URL.Path)

		var req GetMediaUploadURLRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.Equal(t, "trace-123", req.TraceID)
		require.Equal(t, ContentTypeImagePNG, req.ContentType)

		resp := GetMediaUploadURLResponse{
			UploadURL: &mockUploadURL,
			MediaID:   mockMediaID,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(resty.New().SetBaseURL(server.URL))

	request := &GetMediaUploadURLRequest{
		TraceID:       "trace-123",
		ContentType:   ContentTypeImagePNG,
		ContentLength: 1024,
		SHA256Hash:    "abcd1234",
		Field:         "input",
	}

	response, err := client.GetUploadURL(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, mockMediaID, response.MediaID)
	require.NotNil(t, response.UploadURL)
	require.Equal(t, mockUploadURL, *response.UploadURL)
}

func TestClient_GetUploadURL_ValidationError(t *testing.T) {
	client := NewClient(resty.New())

	request := &GetMediaUploadURLRequest{
		// Missing required fields
	}

	_, err := client.GetUploadURL(context.Background(), request)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'traceId' is required")
}

func TestClient_Get(t *testing.T) {
	mockMediaID := "media-123"
	mockTime := time.Now()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, "/media/"+mockMediaID, r.URL.Path)

		resp := GetMediaResponse{
			MediaID:       mockMediaID,
			ContentType:   "image/png",
			ContentLength: 1024,
			UploadedAt:    mockTime,
			URL:           "https://example.com/download",
			URLExpiry:     "2024-01-01T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(resty.New().SetBaseURL(server.URL))

	response, err := client.Get(context.Background(), mockMediaID)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, mockMediaID, response.MediaID)
	require.Equal(t, "image/png", response.ContentType)
	require.Equal(t, 1024, response.ContentLength)
	require.Equal(t, "https://example.com/download", response.URL)
}

func TestClient_Get_EmptyMediaID(t *testing.T) {
	client := NewClient(resty.New())

	_, err := client.Get(context.Background(), "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "'mediaID' is required")
}

func TestClient_Patch(t *testing.T) {
	mockMediaID := "media-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "PATCH", r.Method)
		require.Equal(t, "/media/"+mockMediaID, r.URL.Path)

		var req PatchMediaRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.Equal(t, 200, req.UploadHTTPStatus)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(resty.New().SetBaseURL(server.URL))

	request := &PatchMediaRequest{
		UploadedAt:       time.Now(),
		UploadHTTPStatus: 200,
	}

	err := client.Patch(context.Background(), mockMediaID, request)
	require.NoError(t, err)
}

func TestClient_Patch_ValidationError(t *testing.T) {
	client := NewClient(resty.New())

	request := &PatchMediaRequest{
		// Missing required fields
	}

	err := client.Patch(context.Background(), "media-123", request)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'uploadedAt' is required")
}

func TestClient_Patch_EmptyMediaID(t *testing.T) {
	client := NewClient(resty.New())

	request := &PatchMediaRequest{
		UploadedAt:       time.Now(),
		UploadHTTPStatus: 200,
	}

	err := client.Patch(context.Background(), "", request)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'mediaID' is required")
}

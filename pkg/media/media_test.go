package media

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestGetMediaUploadURLRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		request GetUploadURLRequest
		wantErr bool
		errMsg  string
	}{
		{
			"valid request",
			GetUploadURLRequest{
				TraceID:       "trace-123",
				ContentType:   ContentTypeImagePNG,
				ContentLength: 1024,
				SHA256Hash:    "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
				Field:         "input",
			},
			false,
			"",
		},
		{
			"missing trace id",
			GetUploadURLRequest{
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
			GetUploadURLRequest{
				TraceID:       "trace-123",
				ContentLength: 1024,
				SHA256Hash:    "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
				Field:         "input",
			},
			true,
			"'contentType' is required",
		},
		{
			"invalid content length",
			GetUploadURLRequest{
				TraceID:       "trace-123",
				ContentType:   ContentTypeImagePNG,
				ContentLength: 0,
				SHA256Hash:    "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
				Field:         "input",
			},
			true,
			"'contentLength' must be greater than 0",
		},
		{
			"missing sha256 hash",
			GetUploadURLRequest{
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
			GetUploadURLRequest{
				TraceID:       "trace-123",
				ContentType:   ContentTypeImagePNG,
				ContentLength: 1024,
				SHA256Hash:    "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
			},
			true,
			"'field' is required",
		},
		{
			"invalid field",
			GetUploadURLRequest{
				TraceID:       "trace-123",
				ContentType:   ContentTypeImagePNG,
				ContentLength: 1024,
				SHA256Hash:    "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
				Field:         "invalid",
			},
			true,
			"'field' must be one of: input, output, metadata",
		},
		{
			"invalid sha256 hash length",
			GetUploadURLRequest{
				TraceID:       "trace-123",
				ContentType:   ContentTypeImagePNG,
				ContentLength: 1024,
				SHA256Hash:    "abcd1234",
				Field:         "input",
			},
			true,
			"'sha256Hash' must be a 44 character base64 encoded SHA-256 hash",
		},
		{
			"invalid sha256 hash encoding",
			GetUploadURLRequest{
				TraceID:       "trace-123",
				ContentType:   ContentTypeImagePNG,
				ContentLength: 1024,
				SHA256Hash:    "!@#$%^&*()1234567890abcdefghijklmnopqrstuv!!",
				Field:         "input",
			},
			true,
			"'sha256Hash' must be a valid base64 encoded string",
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

		var req GetUploadURLRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.Equal(t, "trace-123", req.TraceID)
		require.Equal(t, ContentTypeImagePNG, req.ContentType)

		resp := GetUploadURLResponse{
			UploadURL: mockUploadURL,
			MediaID:   mockMediaID,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(resty.New().SetBaseURL(server.URL))

	request := &GetUploadURLRequest{
		TraceID:       "trace-123",
		ContentType:   ContentTypeImagePNG,
		ContentLength: 1024,
		SHA256Hash:    "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
		Field:         "input",
	}

	response, err := client.GetUploadURL(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, mockMediaID, response.MediaID)
	require.NotEmpty(t, response.UploadURL)
	require.Equal(t, mockUploadURL, response.UploadURL)
}

func TestClient_GetUploadURL_ValidationError(t *testing.T) {
	client := NewClient(resty.New())

	request := &GetUploadURLRequest{
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

func TestUploadFromBytesRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		request UploadFromBytesRequest
		wantErr bool
		errMsg  string
	}{
		{
			"valid request",
			UploadFromBytesRequest{
				TraceID:     "trace-123",
				ContentType: ContentTypeImagePNG,
				Field:       "input",
				Data:        []byte("test data"),
			},
			false,
			"",
		},
		{
			"missing trace id",
			UploadFromBytesRequest{
				ContentType: ContentTypeImagePNG,
				Field:       "input",
				Data:        []byte("test data"),
			},
			true,
			"'traceId' is required",
		},
		{
			"missing content type",
			UploadFromBytesRequest{
				TraceID: "trace-123",
				Field:   "input",
				Data:    []byte("test data"),
			},
			true,
			"'contentType' is required",
		},
		{
			"missing field",
			UploadFromBytesRequest{
				TraceID:     "trace-123",
				ContentType: ContentTypeImagePNG,
				Data:        []byte("test data"),
			},
			true,
			"'field' is required",
		},
		{
			"invalid field",
			UploadFromBytesRequest{
				TraceID:     "trace-123",
				ContentType: ContentTypeImagePNG,
				Field:       "invalid",
				Data:        []byte("test data"),
			},
			true,
			"'field' must be one of: input, output, metadata",
		},
		{
			"missing data",
			UploadFromBytesRequest{
				TraceID:     "trace-123",
				ContentType: ContentTypeImagePNG,
				Field:       "input",
			},
			true,
			"'data' is required",
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

func TestUploadFileRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		request UploadFileRequest
		wantErr bool
		errMsg  string
	}{
		{
			"valid request",
			UploadFileRequest{
				TraceID:     "trace-123",
				ContentType: ContentTypeImagePNG,
				Field:       "input",
				FilePath:    "/path/to/file.png",
			},
			false,
			"",
		},
		{
			"missing trace id",
			UploadFileRequest{
				ContentType: ContentTypeImagePNG,
				Field:       "input",
				FilePath:    "/path/to/file.png",
			},
			true,
			"'traceId' is required",
		},
		{
			"missing field",
			UploadFileRequest{
				TraceID:     "trace-123",
				ContentType: ContentTypeImagePNG,
				FilePath:    "/path/to/file.png",
			},
			true,
			"'field' is required",
		},
		{
			"invalid field",
			UploadFileRequest{
				TraceID:     "trace-123",
				ContentType: ContentTypeImagePNG,
				Field:       "invalid",
				FilePath:    "/path/to/file.png",
			},
			true,
			"'field' must be one of: input, output, metadata",
		},
		{
			"missing file path",
			UploadFileRequest{
				TraceID:     "trace-123",
				ContentType: ContentTypeImagePNG,
				Field:       "input",
			},
			true,
			"'filePath' is required",
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

func TestClient_UploadFromBytes(t *testing.T) {
	testData := []byte("test file content")
	hash := sha256.Sum256(testData)
	expectedHash := base64.StdEncoding.EncodeToString(hash[:])
	mockMediaID := "media-123"

	// Mock servers for upload URL and actual upload
	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "PUT", r.Method)
		require.Equal(t, "image/png", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Equal(t, testData, body)

		w.WriteHeader(http.StatusOK)
	}))
	defer uploadServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			require.Equal(t, "/media", r.URL.Path)

			var req GetUploadURLRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			require.Equal(t, "trace-123", req.TraceID)
			require.Equal(t, ContentTypeImagePNG, req.ContentType)
			require.Equal(t, len(testData), req.ContentLength)
			require.Equal(t, expectedHash, req.SHA256Hash)
			require.Equal(t, "input", req.Field)

			resp := GetUploadURLResponse{
				UploadURL: uploadServer.URL + "/upload",
				MediaID:   mockMediaID,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		case "PATCH":
			require.Equal(t, "/media/"+mockMediaID, r.URL.Path)

			var req PatchMediaRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			require.Equal(t, 200, req.UploadHTTPStatus)
			require.Empty(t, req.UploadHTTPError)
			require.GreaterOrEqual(t, req.UploadTimeMs, 0)

			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer apiServer.Close()

	client := NewClient(resty.New().SetBaseURL(apiServer.URL))

	request := &UploadFromBytesRequest{
		TraceID:     "trace-123",
		ContentType: ContentTypeImagePNG,
		Field:       "input",
		Data:        testData,
	}

	response, err := client.UploadFromBytes(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, mockMediaID, response.MediaID)
}

func TestClient_UploadFromBytes_ValidationError(t *testing.T) {
	client := NewClient(resty.New())

	request := &UploadFromBytesRequest{
		// Missing required fields
	}

	_, err := client.UploadFromBytes(context.Background(), request)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'traceId' is required")
}

func TestClient_UploadFromBytes_ExistingFile(t *testing.T) {
	mockMediaID := "media-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "/media", r.URL.Path)

		// Return empty upload URL to simulate existing file
		resp := GetUploadURLResponse{
			UploadURL: "",
			MediaID:   mockMediaID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(resty.New().SetBaseURL(server.URL))

	request := &UploadFromBytesRequest{
		TraceID:     "trace-123",
		ContentType: ContentTypeImagePNG,
		Field:       "input",
		Data:        []byte("test data"),
	}

	response, err := client.UploadFromBytes(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, mockMediaID, response.MediaID)
}

func TestClient_UploadFile(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.png")
	testData := []byte("test file content")
	err := os.WriteFile(testFile, testData, 0644)
	require.NoError(t, err)

	hash := sha256.Sum256(testData)
	expectedHash := base64.StdEncoding.EncodeToString(hash[:])
	mockMediaID := "media-123"

	// Mock servers for upload URL and actual upload
	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "PUT", r.Method)
		require.Equal(t, "image/png", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer uploadServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			require.Equal(t, "/media", r.URL.Path)

			var req GetUploadURLRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			require.Equal(t, "trace-123", req.TraceID)
			require.Equal(t, ContentTypeImagePNG, req.ContentType)
			require.Equal(t, len(testData), req.ContentLength)
			require.Equal(t, expectedHash, req.SHA256Hash)
			require.Equal(t, "input", req.Field)

			resp := GetUploadURLResponse{
				UploadURL: uploadServer.URL + "/upload",
				MediaID:   mockMediaID,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		case "PATCH":
			require.Equal(t, "/media/"+mockMediaID, r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer apiServer.Close()

	client := NewClient(resty.New().SetBaseURL(apiServer.URL))

	request := &UploadFileRequest{
		TraceID:     "trace-123",
		ContentType: ContentTypeImagePNG,
		Field:       "input",
		FilePath:    testFile,
	}

	response, err := client.UploadFile(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, mockMediaID, response.MediaID)
}

func TestClient_UploadFile_ValidationError(t *testing.T) {
	client := NewClient(resty.New())

	request := &UploadFileRequest{
		// Missing required fields
	}

	_, err := client.UploadFile(context.Background(), request)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'traceId' is required")
}

func TestClient_UploadFile_FileNotFound(t *testing.T) {
	client := NewClient(resty.New())

	request := &UploadFileRequest{
		TraceID:     "trace-123",
		ContentType: ContentTypeImagePNG,
		Field:       "input",
		FilePath:    "/nonexistent/file.png",
	}

	_, err := client.UploadFile(context.Background(), request)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read file")
}

func TestClient_UploadFile_AutoDetectContentType(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.jpg")
	testData := []byte("test image data")
	err := os.WriteFile(testFile, testData, 0644)
	require.NoError(t, err)

	mockMediaID := "media-123"

	// Mock servers
	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "PUT", r.Method)
		require.Equal(t, "image/jpeg", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer uploadServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			var req GetUploadURLRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			require.Equal(t, ContentType("image/jpeg"), req.ContentType) // Auto-detected from .jpg extension

			resp := GetUploadURLResponse{
				UploadURL: uploadServer.URL + "/upload",
				MediaID:   mockMediaID,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		case "PATCH":
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer apiServer.Close()

	client := NewClient(resty.New().SetBaseURL(apiServer.URL))

	request := &UploadFileRequest{
		TraceID:  "trace-123",
		Field:    "input",
		FilePath: testFile,
		// ContentType not specified - should be auto-detected
	}

	response, err := client.UploadFile(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, mockMediaID, response.MediaID)
}

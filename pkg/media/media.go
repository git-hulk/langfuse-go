// Package media provides functionality for managing media records in Langfuse.
//
// This package allows you to upload, retrieve, and manage media files associated
// with traces and observations. Media files can include images, audio, video,
// documents, and other file types supported by the Langfuse platform.
package media

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// ContentType represents supported MIME types for media records.
type ContentType string

const (
	ContentTypeImagePNG               ContentType = "image/png"
	ContentTypeImageJPEG              ContentType = "image/jpeg"
	ContentTypeImageJPG               ContentType = "image/jpg"
	ContentTypeImageWebP              ContentType = "image/webp"
	ContentTypeImageGIF               ContentType = "image/gif"
	ContentTypeImageSVGXML            ContentType = "image/svg+xml"
	ContentTypeImageTIFF              ContentType = "image/tiff"
	ContentTypeImageBMP               ContentType = "image/bmp"
	ContentTypeAudioMPEG              ContentType = "audio/mpeg"
	ContentTypeAudioMP3               ContentType = "audio/mp3"
	ContentTypeAudioWAV               ContentType = "audio/wav"
	ContentTypeAudioOGG               ContentType = "audio/ogg"
	ContentTypeAudioOGA               ContentType = "audio/oga"
	ContentTypeAudioAAC               ContentType = "audio/aac"
	ContentTypeAudioMP4               ContentType = "audio/mp4"
	ContentTypeAudioFLAC              ContentType = "audio/flac"
	ContentTypeVideoMP4               ContentType = "video/mp4"
	ContentTypeVideoWebM              ContentType = "video/webm"
	ContentTypeTextPlain              ContentType = "text/plain"
	ContentTypeTextHTML               ContentType = "text/html"
	ContentTypeTextCSS                ContentType = "text/css"
	ContentTypeTextCSV                ContentType = "text/csv"
	ContentTypeApplicationPDF         ContentType = "application/pdf"
	ContentTypeApplicationMSWord      ContentType = "application/msword"
	ContentTypeApplicationMSExcel     ContentType = "application/vnd.ms-excel"
	ContentTypeApplicationZIP         ContentType = "application/zip"
	ContentTypeApplicationJSON        ContentType = "application/json"
	ContentTypeApplicationXML         ContentType = "application/xml"
	ContentTypeApplicationOctetStream ContentType = "application/octet-stream"
)

// GetUploadURLRequest represents the request to get a presigned upload URL for media.
type GetUploadURLRequest struct {
	TraceID       string      `json:"traceId"`
	ObservationID string      `json:"observationId,omitempty"`
	ContentType   ContentType `json:"contentType"`
	ContentLength int         `json:"contentLength"`
	SHA256Hash    string      `json:"sha256Hash"`
	Field         string      `json:"field"`
}

func (r *GetUploadURLRequest) validate() error {
	if r.TraceID == "" {
		return errors.New("'traceId' is required")
	}
	if r.ContentType == "" {
		return errors.New("'contentType' is required")
	}
	if r.ContentLength <= 0 {
		return errors.New("'contentLength' must be greater than 0")
	}
	if r.SHA256Hash == "" {
		return errors.New("'sha256Hash' is required")
	}
	if len(r.SHA256Hash) != 44 {
		return errors.New("'sha256Hash' must be a 44 character base64 encoded SHA-256 hash")
	}
	if _, err := base64.StdEncoding.DecodeString(r.SHA256Hash); err != nil {
		return errors.New("'sha256Hash' must be a valid base64 encoded string")
	}
	if r.Field == "" {
		return errors.New("'field' is required")
	}
	if r.Field != "input" && r.Field != "output" && r.Field != "metadata" {
		return fmt.Errorf("'field' must be one of: input, output, metadata")
	}
	return nil
}

// GetUploadURLResponse represents the response from getting a presigned upload URL.
type GetUploadURLResponse struct {
	UploadURL string `json:"uploadUrl,omitempty"`
	MediaID   string `json:"mediaId"`
}

// GetMediaResponse represents the response from getting a media record.
type GetMediaResponse struct {
	MediaID       string    `json:"mediaId"`
	ContentType   string    `json:"contentType"`
	ContentLength int       `json:"contentLength"`
	UploadedAt    time.Time `json:"uploadedAt"`
	URL           string    `json:"url"`
	URLExpiry     string    `json:"urlExpiry"`
}

// PatchMediaRequest represents the request to update a media record.
type PatchMediaRequest struct {
	UploadedAt       time.Time `json:"uploadedAt"`
	UploadHTTPStatus int       `json:"uploadHttpStatus"`
	UploadHTTPError  string    `json:"uploadHttpError,omitempty"`
	UploadTimeMs     int       `json:"uploadTimeMs,omitempty"`
}

func (r *PatchMediaRequest) validate() error {
	if r.UploadedAt.IsZero() {
		return errors.New("'uploadedAt' is required")
	}
	return nil
}

// Client represents the media API client.
type Client struct {
	restyCli *resty.Client
}

// NewClient creates a new media API client.
func NewClient(cli *resty.Client) *Client {
	return &Client{restyCli: cli}
}

// GetUploadURL retrieves a presigned upload URL for uploading media.
//
// This endpoint returns a presigned URL that can be used to upload media files
// directly to the storage provider. If the media file is already uploaded
// (based on SHA256 hash), the upload URL will be null.
func (c *Client) GetUploadURL(ctx context.Context, request *GetUploadURLRequest) (*GetUploadURLResponse, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	var response GetUploadURLResponse
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(request).
		SetResult(&response).
		Post("/media")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("get media upload url failed with status code %d", rsp.StatusCode())
	}
	return &response, nil
}

// Get retrieves a specific media record by ID.
//
// Returns the media record metadata including content type, size, upload date,
// and a download URL with expiry information.
func (c *Client) Get(ctx context.Context, mediaID string) (*GetMediaResponse, error) {
	if mediaID == "" {
		return nil, errors.New("'mediaID' is required")
	}

	var media GetMediaResponse
	req := c.restyCli.R().
		SetContext(ctx).
		SetResult(&media).
		SetPathParam("mediaId", mediaID)

	rsp, err := req.Get("/media/{mediaId}")
	if err != nil {
		return nil, err
	}
	if rsp.IsError() {
		return nil, fmt.Errorf("get media failed with status code %d", rsp.StatusCode())
	}
	return &media, nil
}

// Patch updates a media record with upload status information.
//
// This endpoint is typically used to report the status of a media upload
// after using the presigned URL obtained from GetUploadURL.
func (c *Client) Patch(ctx context.Context, mediaID string, request *PatchMediaRequest) error {
	if mediaID == "" {
		return errors.New("'mediaID' is required")
	}
	if err := request.validate(); err != nil {
		return err
	}

	req := c.restyCli.R().
		SetContext(ctx).
		SetBody(request).
		SetPathParam("mediaId", mediaID)

	rsp, err := req.Patch("/media/{mediaId}")
	if err != nil {
		return err
	}
	if rsp.IsError() {
		return fmt.Errorf("patch media failed with status code %d", rsp.StatusCode())
	}
	return nil
}

// UploadFromBytesRequest represents the request for uploading media from bytes.
type UploadFromBytesRequest struct {
	TraceID       string      `json:"traceId"`
	ObservationID string      `json:"observationId,omitempty"`
	ContentType   ContentType `json:"contentType"`
	Field         string      `json:"field"`
	Data          []byte      `json:"-"` // Not serialized to JSON
}

func (r *UploadFromBytesRequest) validate() error {
	if r.TraceID == "" {
		return errors.New("'traceId' is required")
	}
	if r.ContentType == "" {
		return errors.New("'contentType' is required")
	}
	if r.Field == "" {
		return errors.New("'field' is required")
	}
	if r.Field != "input" && r.Field != "output" && r.Field != "metadata" {
		return fmt.Errorf("'field' must be one of: input, output, metadata")
	}
	if len(r.Data) == 0 {
		return errors.New("'data' is required")
	}
	return nil
}

// UploadFileRequest represents the request for uploading a media file.
type UploadFileRequest struct {
	TraceID       string      `json:"traceId"`
	ObservationID string      `json:"observationId,omitempty"`
	ContentType   ContentType `json:"contentType"`
	Field         string      `json:"field"`
	FilePath      string      `json:"-"` // Not serialized to JSON
}

func (r *UploadFileRequest) validate() error {
	if r.TraceID == "" {
		return errors.New("'traceId' is required")
	}
	if r.Field == "" {
		return errors.New("'field' is required")
	}
	if r.Field != "input" && r.Field != "output" && r.Field != "metadata" {
		return fmt.Errorf("'field' must be one of: input, output, metadata")
	}
	if r.FilePath == "" {
		return errors.New("'filePath' is required")
	}
	return nil
}

// UploadResponse represents the response from uploading media.
type UploadResponse struct {
	MediaID string `json:"mediaId"`
}

// UploadFromBytes uploads media from byte data.
//
// This method handles the complete upload flow: getting a presigned URL,
// uploading the data, and updating the media record with upload status.
func (c *Client) UploadFromBytes(ctx context.Context, request *UploadFromBytesRequest) (*UploadResponse, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	// Calculate SHA256 hash
	hash := sha256.Sum256(request.Data)
	sha256Hash := base64.StdEncoding.EncodeToString(hash[:])

	// Get upload URL
	uploadURLReq := &GetUploadURLRequest{
		TraceID:       request.TraceID,
		ObservationID: request.ObservationID,
		ContentType:   request.ContentType,
		ContentLength: len(request.Data),
		SHA256Hash:    sha256Hash,
		Field:         request.Field,
	}

	uploadURLRsp, err := c.GetUploadURL(ctx, uploadURLReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get upload URL: %w", err)
	}

	// If upload URL is empty, file already exists
	if uploadURLRsp.UploadURL == "" {
		return &UploadResponse{MediaID: uploadURLRsp.MediaID}, nil
	}

	startTime := time.Now()
	uploadRsp, err := resty.New().R().
		SetContext(ctx).
		SetHeader("Content-Type", string(request.ContentType)).
		SetHeader("x-amz-checksum-sha256", sha256Hash).
		SetBody(request.Data).
		Put(uploadURLRsp.UploadURL)

	uploadTimeMs := int(time.Since(startTime).Milliseconds())

	// Update media record with upload status
	patchReq := &PatchMediaRequest{
		UploadedAt:   time.Now(),
		UploadTimeMs: uploadTimeMs,
	}

	if err != nil {
		patchReq.UploadHTTPStatus = 0 // Use 0 for network errors
		patchReq.UploadHTTPError = err.Error()
	} else {
		patchReq.UploadHTTPStatus = uploadRsp.StatusCode()
		if uploadRsp.IsError() {
			patchReq.UploadHTTPError = fmt.Sprintf("HTTP %d: %s", uploadRsp.StatusCode(), uploadRsp.String())
		}
	}

	if patchErr := c.Patch(ctx, uploadURLRsp.MediaID, patchReq); patchErr != nil {
		return nil, fmt.Errorf("failed to update media record: %w", patchErr)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to upload media: %w", err)
	}
	if uploadRsp.IsError() {
		return nil, fmt.Errorf("upload failed with status %d: %s", uploadRsp.StatusCode(), uploadRsp.String())
	}

	return &UploadResponse{MediaID: uploadURLRsp.MediaID}, nil
}

func getContentTypeFromFileExtension(filePath string) (ContentType, error) {
	ext := filepath.Ext(filePath)
	// For some extensions, mime.TypeByExtension may return multiple types separated by semicolon.
	// For example: "text/html; charset=utf-8", and we only need the first part.
	fields := strings.Split(mime.TypeByExtension(ext), ";")
	mimeType := strings.TrimSpace(fields[0])
	if mimeType != "" {
		return ContentType(mimeType), nil
	}
	return "", fmt.Errorf("could not determine content type for file extension %s", ext)
}

// UploadFile uploads a media file from the local filesystem.
//
// This method reads the file from the provided path and uploads it using UploadFromBytes.
// If no content type is specified, it will be auto-detected from the file extension.
func (c *Client) UploadFile(ctx context.Context, request *UploadFileRequest) (*UploadResponse, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	// Read file data
	data, err := os.ReadFile(request.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Auto-detect content type if not specified
	contentType := request.ContentType
	if contentType == "" {
		contentType, err = getContentTypeFromFileExtension(request.FilePath)
		if err != nil {
			return nil, err
		}
	}

	return c.UploadFromBytes(ctx, &UploadFromBytesRequest{
		TraceID:       request.TraceID,
		ObservationID: request.ObservationID,
		ContentType:   contentType,
		Field:         request.Field,
		Data:          data,
	})
}

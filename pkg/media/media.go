// Package media provides functionality for managing media records in Langfuse.
//
// This package allows you to upload, retrieve, and manage media files associated
// with traces and observations. Media files can include images, audio, video,
// documents, and other file types supported by the Langfuse platform.
package media

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// MediaContentType represents supported MIME types for media records.
type MediaContentType string

const (
	ContentTypeImagePNG               MediaContentType = "image/png"
	ContentTypeImageJPEG              MediaContentType = "image/jpeg"
	ContentTypeImageJPG               MediaContentType = "image/jpg"
	ContentTypeImageWebP              MediaContentType = "image/webp"
	ContentTypeImageGIF               MediaContentType = "image/gif"
	ContentTypeImageSVGXML            MediaContentType = "image/svg+xml"
	ContentTypeImageTIFF              MediaContentType = "image/tiff"
	ContentTypeImageBMP               MediaContentType = "image/bmp"
	ContentTypeAudioMPEG              MediaContentType = "audio/mpeg"
	ContentTypeAudioMP3               MediaContentType = "audio/mp3"
	ContentTypeAudioWAV               MediaContentType = "audio/wav"
	ContentTypeAudioOGG               MediaContentType = "audio/ogg"
	ContentTypeAudioOGA               MediaContentType = "audio/oga"
	ContentTypeAudioAAC               MediaContentType = "audio/aac"
	ContentTypeAudioMP4               MediaContentType = "audio/mp4"
	ContentTypeAudioFLAC              MediaContentType = "audio/flac"
	ContentTypeVideoMP4               MediaContentType = "video/mp4"
	ContentTypeVideoWebM              MediaContentType = "video/webm"
	ContentTypeTextPlain              MediaContentType = "text/plain"
	ContentTypeTextHTML               MediaContentType = "text/html"
	ContentTypeTextCSS                MediaContentType = "text/css"
	ContentTypeTextCSV                MediaContentType = "text/csv"
	ContentTypeApplicationPDF         MediaContentType = "application/pdf"
	ContentTypeApplicationMSWord      MediaContentType = "application/msword"
	ContentTypeApplicationMSExcel     MediaContentType = "application/vnd.ms-excel"
	ContentTypeApplicationZIP         MediaContentType = "application/zip"
	ContentTypeApplicationJSON        MediaContentType = "application/json"
	ContentTypeApplicationXML         MediaContentType = "application/xml"
	ContentTypeApplicationOctetStream MediaContentType = "application/octet-stream"
)

// GetMediaUploadURLRequest represents the request to get a presigned upload URL for media.
type GetMediaUploadURLRequest struct {
	TraceID       string           `json:"traceId"`
	ObservationID *string          `json:"observationId,omitempty"`
	ContentType   MediaContentType `json:"contentType"`
	ContentLength int              `json:"contentLength"`
	SHA256Hash    string           `json:"sha256Hash"`
	Field         string           `json:"field"`
}

func (r *GetMediaUploadURLRequest) validate() error {
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
	if r.Field == "" {
		return errors.New("'field' is required")
	}
	if r.Field != "input" && r.Field != "output" && r.Field != "metadata" {
		return fmt.Errorf("'field' must be one of: input, output, metadata")
	}
	return nil
}

// GetMediaUploadURLResponse represents the response from getting a presigned upload URL.
type GetMediaUploadURLResponse struct {
	UploadURL *string `json:"uploadUrl,omitempty"`
	MediaID   string  `json:"mediaId"`
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
	UploadHTTPError  *string   `json:"uploadHttpError,omitempty"`
	UploadTimeMs     *int      `json:"uploadTimeMs,omitempty"`
}

func (r *PatchMediaRequest) validate() error {
	if r.UploadedAt.IsZero() {
		return errors.New("'uploadedAt' is required")
	}
	if r.UploadHTTPStatus == 0 {
		return errors.New("'uploadHttpStatus' is required")
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
func (c *Client) GetUploadURL(ctx context.Context, request *GetMediaUploadURLRequest) (*GetMediaUploadURLResponse, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	var response GetMediaUploadURLResponse
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

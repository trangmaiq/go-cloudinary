package cloudinary

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// UploadService handles communication with the uploading related

type UploadService service

type UploadRequest struct {
	// Required fields to call upload request
	File         string `json:"file"`
	UploadPreset string `json:"upload_preset"`
	Timestamp    string `json:"timestamp"`

	// Optional fields to call upload request
	// Naming and storage fields
	PublicId       string `json:"public_id,omitempty"`
	Folder         string `json:"folder,omitempty"`
	UseFilename    bool   `json:"use_filename,omitempty"`
	UniqueFilename bool   `json:"unique_filename,omitempty"`
	ResourceType   string `json:"resource_type,omitempty"`
	Type           string `json:"type,omitempty"`
	//AccessControl *AccessControl `json:"access_control"
	AccessMode              string `json:"access_mode,omitempty"`
	DiscardOriginalFilename bool   `json:"discard_original_filename,omitempty"`
	Overwrite               bool   `json:"overwrite,omitempty"`

	// Resource data fields
	Tags                  string      `json:"tags,omitempty"`
	Context               string      `json:"context,omitempty"`
	Colors                bool        `json:"colors,omitempty"`
	Faces                 bool        `json:"faces,omitempty"`
	QualityAnalysis       bool        `json:"quality_analysis,omitempty"`
	ImageMetadata         bool        `json:"image_metadata,omitempty"`
	Phash                 bool        `json:"phash,omitempty"`
	ResponsiveBreakpoints interface{} `json:"responsive_breakpoints,omitempty"`
	AutoTagging           float64     `json:"auto_tagging,omitempty"`
	Categorization        string      `json:"categorization,omitempty"`
	Detection             string      `json:"detection,omitempty"`
	ORC                   string      `json:"ocr,omitempty"`
	Exif                  bool        `json:"exif,omitempty"`

	// Manipulations fields
	Eager                string `json:"eager,omitempty"`
	EagerAsync           bool   `json:"eager_async,omitempty"`
	EagerNotificationURL string `json:"eager_notification_url,omitempty"`
	Transformation       string `json:"transformation,omitempty"`
	Format               string `json:"format,omitempty"`
	CustomCoordinates    string `json:"custom_coordinates,omitempty"`
	FaceCoordinates      string `json:"face_coordinates,omitempty"`
	BackgroundRemoval    string `json:"background_removal,omitempty"`
	RawConvert           string `json:"raw_convert,omitempty"`

	// Additional options fields
	AllowedFormats    string `json:"allowed_formats,omitempty"`
	Async             bool   `json:"async,omitempty"`
	Backup            bool   `json:"backup,omitempty"`
	Callback          string `json:"callback,omitempty"`
	Headers           string `json:"headers,omitempty"`
	Invalidate        bool   `json:"invalidate,omitempty"`
	Moderation        string `json:"moderation,omitempty"`
	Proxy             string `json:"proxy,omitempty"`
	ReturnDeleteToken bool   `json:"return_delete_token,omitempty"`
}

type UploadResponse struct {
	PublicId         string   `json:"public_id"`
	Version          int64    `json:"version"`
	Signature        string   `json:"signature"`
	Width            int64    `json:"width"`
	Height           int64    `json:"height"`
	Format           string   `json:"format"`
	ResourceType     string   `json:"resource_type"`
	CreatedAt        string   `json:"created_at"`
	Tags             []string `json:"tags"`
	Bytes            int64    `json:"bytes"`
	Type             string   `json:"type"`
	Etag             string   `json:"etag"`
	Placeholder      bool     `json:"placeholder"`
	URL              string   `json:"url"`
	SecureURL        string   `json:"secure_url"`
	AccessMode       string   `json:"access_mode"`
	OriginalFilename string   `json:"original_filename"`
}

func (us *UploadService) UploadImage(ctx context.Context, request *UploadRequest) (*UploadResponse, *Response, error) {
	us.prepareUploadImage(request)
	u := fmt.Sprintf("image/upload")

	timeStamp := strconv.Itoa(int(time.Now().UTC().Unix())) + us.client.apiSecret
	request.Timestamp = timeStamp

	req, err := us.client.NewRequest("POST", u, request)
	if err != nil {
		return nil, nil, err
	}

	ur := new(UploadResponse)
	resp, err := us.client.Do(ctx, req, ur)
	if err != nil {
		return nil, resp, err
	}

	return ur, resp, nil
}

func (us *UploadService) prepareUploadImage(request *UploadRequest) {
	switch {
	case strings.HasPrefix(request.File, "/"):
	// Upload image using local path
	case strings.HasPrefix(request.File, "s3"):
	// Upload image using Amazon S3
	case strings.HasPrefix(request.File, "gs"):
		// Upload image using Google Storage
	default:
		// Upload image using HTTPS URL or HTTP
	}
}

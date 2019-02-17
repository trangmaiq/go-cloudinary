package cloudinary

import (
	"context"
	"fmt"
	"strconv"
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
	Tags            string `json:"tags,omitempty"`
	Context         string `json:"context,omitempty"`
	Colors          bool   `json:"colors,omitempty"`
	Faces           bool   `json:"faces,omitempty"`
	QualityAnalysis bool   `json:"quality_analysis,omitempty"`
	ImageMetadata   bool   `json:"image_metadata,omitempty"`
	Phash           bool   `json:"phash,omitempty"`
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

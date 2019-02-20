package cloudinary

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
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
}

type UploadOptions struct {
	// Optional fields to call upload request
	// Naming and storage fields
	PublicId       *string `json:"public_id,omitempty"`
	Folder         *string `json:"folder,omitempty"`
	UseFilename    *bool   `json:"use_filename,omitempty"`
	UniqueFilename *bool   `json:"unique_filename,omitempty"`
	ResourceType   *string `json:"resource_type,omitempty"`
	Type           *string `json:"type,omitempty"`
	//AccessControl           interface{} `json:"access_control,omitempty"`
	AccessMode              *string `json:"access_mode,omitempty"`
	DiscardOriginalFilename *bool   `json:"discard_original_filename,omitempty"`
	Overwrite               *bool   `json:"overwrite,omitempty"`

	// Resource data fields
	Tags            *string `json:"tags,omitempty"`
	Context         *string `json:"context,omitempty"`
	Colors          *bool   `json:"colors,omitempty"`
	Faces           *bool   `json:"faces,omitempty"`
	QualityAnalysis *bool   `json:"quality_analysis,omitempty"`
	ImageMetadata   *bool   `json:"image_metadata,omitempty"`
	Phash           *bool   `json:"phash,omitempty"`
	//ResponsiveBreakpoints interface{} `json:"responsive_breakpoints,omitempty"`
	AutoTagging    *float64 `json:"auto_tagging,omitempty"`
	Categorization *string  `json:"categorization,omitempty"`
	Detection      *string  `json:"detection,omitempty"`
	ORC            *string  `json:"ocr,omitempty"`
	Exif           *bool    `json:"exif,omitempty"`

	// Manipulations fields
	Eager                *string `json:"eager,omitempty"`
	EagerAsync           *bool   `json:"eager_async,omitempty"`
	EagerNotificationURL *string `json:"eager_notification_url,omitempty"`
	Transformation       *string `json:"transformation,omitempty"`
	Format               *string `json:"format,omitempty"`
	CustomCoordinates    *string `json:"custom_coordinates,omitempty"`
	FaceCoordinates      *string `json:"face_coordinates,omitempty"`
	BackgroundRemoval    *string `json:"background_removal,omitempty"`
	RawConvert           *string `json:"raw_convert,omitempty"`

	// Additional options fields
	AllowedFormats    *string `json:"allowed_formats,omitempty"`
	Async             *bool   `json:"async,omitempty"`
	Backup            *bool   `json:"backup,omitempty"`
	Callback          *string `json:"callback,omitempty"`
	Headers           *string `json:"headers,omitempty"`
	Invalidate        *bool   `json:"invalidate,omitempty"`
	Moderation        *string `json:"moderation,omitempty"`
	Proxy             *string `json:"proxy,omitempty"`
	ReturnDeleteToken *bool   `json:"return_delete_token,omitempty"`
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

type Opt func(uo *UploadOptions)

func WithPublicId(id string) Opt {
	return func(uo *UploadOptions) {
		uo.PublicId = &id
	}
}

func WithFolder(folder string) Opt {
	return func(uo *UploadOptions) {
		uo.Folder = &folder
	}
}

func WithUseFilename(isUseFilename bool) Opt {
	return func(uo *UploadOptions) {
		uo.UseFilename = &isUseFilename
	}
}

func WithUniqueFilename(isUniqueFilename bool) Opt {
	return func(uo *UploadOptions) {
		uo.UniqueFilename = &isUniqueFilename
	}
}

func WithResourceType(resourceType string) Opt {
	return func(uo *UploadOptions) {
		uo.ResourceType = &resourceType
	}
}

func WithType(typeStr string) Opt {
	return func(uo *UploadOptions) {
		uo.Type = &typeStr
	}
}

func WithAccessMode(accessMode string) Opt {
	return func(uo *UploadOptions) {
		uo.AccessMode = &accessMode
	}
}

func WithDiscardOriginalFilename(dof bool) Opt {
	return func(uo *UploadOptions) {
		uo.DiscardOriginalFilename = &dof
	}
}

func WithOverwrite(isOverwrite bool) Opt {
	return func(uo *UploadOptions) {
		uo.Overwrite = &isOverwrite
	}
}

func (us *UploadService) UploadImage(ctx context.Context, request *UploadRequest, opts ...Opt) (*UploadResponse, *Response, error) {
	u := fmt.Sprintf("image/upload")

	opt := new(UploadOptions)
	for _, o := range opts {
		o(opt)
	}

	switch {
	case strings.HasPrefix(request.File, "/"):
		// Upload image using local path
		return us.uploadFromLocalPath(ctx, u, request, opt)
	case strings.HasPrefix(request.File, "s3"):
		// Upload image using Amazon S3
		return us.uploadFromS3(ctx, u, request, opt)
	case strings.HasPrefix(request.File, "gs"):
		// Upload image using Google Storage
		return us.uploadFromGoogleStorage(ctx, u, request, opt)

	default:
		// Upload image using HTTPS URL or HTTP
		return us.uploadFromURL(ctx, u, request, opt)
	}
}

func (us *UploadService) uploadFromURL(ctx context.Context, url string, request *UploadRequest, opt *UploadOptions) (*UploadResponse, *Response, error) {
	req, err := us.client.NewRequest("POST", url, request)
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

func (us *UploadService) uploadFromLocalPath(ctx context.Context, url string, request *UploadRequest, opt *UploadOptions) (*UploadResponse, *Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if request != nil {
		if err := us.buildParamsFromRequest(request, writer); err != nil {
			return nil, nil, err
		}
	}

	if opt != nil {
		if err := us.buildParamsFromOption(opt, writer); err != nil {
			return nil, nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, nil, err
	}

	req, err := us.client.NewUploadRequest(url, body, writer)

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

func (us *UploadService) buildParamsFromRequest(request *UploadRequest, writer *multipart.Writer) error {
	timeStamp := strconv.Itoa(int(time.Now().UTC().Unix())) + us.client.apiSecret
	if err := writer.WriteField("timestamp", timeStamp); err != nil {
		return err
	}

	if err := writer.WriteField("upload_preset", request.UploadPreset); err != nil {
		return err
	}

	file, _, err := us.openFile(request.File)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if stat.IsDir() {
		return errors.New("the asset to upload can't be a directory")
	}

	part, err := writer.CreateFormFile("file", file.Name())
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	return err
}

func (us *UploadService) buildParamsFromOption(opt *UploadOptions, writer *multipart.Writer) error {

	var optMap map[string]interface{}
	optByte, _ := json.Marshal(opt)
	err := json.Unmarshal(optByte, &optMap)
	if err != nil {
		return err
	}

	for field, val := range optMap {
		valStr := fmt.Sprintf("%v", val)
		err := writer.WriteField(field, valStr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (us *UploadService) openFile(filePath string) (file *os.File, dir string, err error) {
	dir, err = os.Getwd()
	if err != nil {
		return nil, dir, err
	}
	file, err = os.Open(dir + filePath)
	return file, dir, err

}

func (us *UploadService) uploadFromS3(ctx context.Context, url string, request *UploadRequest, opt *UploadOptions) (*UploadResponse, *Response, error) {
	return &UploadResponse{}, &Response{}, nil
}

func (us *UploadService) uploadFromGoogleStorage(ctx context.Context, url string, request *UploadRequest, opt *UploadOptions) (*UploadResponse, *Response, error) {
	return &UploadResponse{}, &Response{}, nil
}

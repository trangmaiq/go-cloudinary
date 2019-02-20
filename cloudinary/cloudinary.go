package cloudinary

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-querystring/query"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

const (
	defaultBaseURL = "https://api.cloudinary.com/v1_1/"
)

// A Client manages communication with the Cloudinary API
type Client struct {
	client *http.Client // HTTP client used to communicate with the API

	// BaseURL for the API request. It should always be specified with a trailing slash
	BaseURL *url.URL

	common service // Reuse a single struct instead of allocating one of each service on the heap

	apiKey    string // The API key required to call Cloudinary API
	apiSecret string // The secret key required to sign the token
	cloudName string

	// Services used for talking to different parts of the Cloudinary API
	Upload *UploadService
}

type service struct {
	client *Client
}

// NewClient returns a new Cloudinary API client.
// If a nil httpClient is provided, http.DefaultClient will be used.

// The uri parameter must be a valid URI with the cloudinary:// scheme,
// e.g.
// cloudinary://api_key:api_secret@cloud_name
func NewClient(httpClient *http.Client, uri string) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "cloudinary" {
		return nil, errors.New("missing cloudinary:// scheme in URI")
	}

	secret, exists := u.User.Password()
	if !exists {
		return nil, errors.New("No API secret provided in URI.")
	}

	baseURLStr := defaultBaseURL + u.Host + "/"
	baseURL, _ := url.Parse(baseURLStr)

	c := &Client{
		client:    httpClient,
		BaseURL:   baseURL,
		apiKey:    u.User.Username(),
		apiSecret: secret,
	}
	c.common.client = c
	c.Upload = (*UploadService)(&c.common)

	return c, nil
}

// NewRequest create an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash.
// If specified, the value pointed to by body is JSON encoded
// and included as the request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}
	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	//req.Header.Set("Accept", mediaTypeV3)
	//if c.UserAgent != "" {
	//	req.Header.Set("User-Agent", c.UserAgent)
	//}
	return req, nil
}

func (c *Client) NewUploadRequest(urlStr string, reader io.Reader, writer *multipart.Writer) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}
	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", u.String(), reader)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, err
}

// Response is a Cloudinary API response.
type Response struct {
	*http.Response
}

// newResponse creates a new Response for the provided http.Response.
// r must not be nil.
func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	return response
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v,
// or returned as an error if an API error has occurred.
//
// The provided ctx must be non-nil. If it is canceled or times out,
// ctx.Err() will be returned.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	req = withContext(ctx, req)

	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// If the error type is *url.Error, sanitize its URL before returning.
		if e, ok := err.(*url.Error); ok {
			if url, err := url.Parse(e.URL); err == nil {
				e.URL = sanitizeURL(url).String()
				return nil, e
			}
		}

		return nil, err
	}
	defer resp.Body.Close()

	response := newResponse(resp)
	err = CheckResponse(resp)
	if err != nil {
		return response, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			decErr := json.NewDecoder(resp.Body).Decode(v)
			if decErr == io.EOF {
				decErr = nil // ignore EOF errors caused by empty response body
			}
			if decErr != nil {
				err = decErr
			}
		}
	}

	return response, err
}

func withContext(ctx context.Context, req *http.Request) *http.Request {
	return req.WithContext(ctx)
}

// sanitizeURL redacts the client_secret parameter from the URL which may be
// exposed to the user.
func sanitizeURL(uri *url.URL) *url.URL {
	if uri == nil {
		return nil
	}
	params := uri.Query()
	if len(params.Get("client_secret")) > 0 {
		params.Set("client_secret", "REDACTED")
		uri.RawQuery = params.Encode()
	}
	return uri
}

type ErrorResponse struct {
	// HTTP response that caused this error
	Response *http.Response

	ErrorData struct {
		Message string `json:"message"`
	} `json:"error"`

	// Errors can also include a documentation_url field
	// pointing to some content that might help you resolve the error
	DocumentationURL string `json:"documentation_url,omitempty"`
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, sanitizeURL(r.Response.Request.URL),
		r.Response.StatusCode, r.ErrorData.Message)
}

// CheckResponse checks the API response for errors, and returns them if
// present. A response is considered an error if it has a status code outside
// the 200 range or equal to 202 Accepted.
// API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse. Any other
// response body will be silently ignored.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		if err := json.Unmarshal(data, errorResponse); err != nil {
			return err
		}
	}
	switch {
	default:
		return errorResponse
	}
}

// addOptions adds the parameters in opt as URL query parameters to s. opt
// must be a struct whose fields may contain "url" tags.
func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

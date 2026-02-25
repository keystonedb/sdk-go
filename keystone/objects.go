package keystone

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/keystonedb/sdk-go/proto"
)

type EntityObject struct {
	storageClass       proto.ObjectType
	path               string
	uploadURL          string
	public             bool
	expiry             time.Time
	contentType        string
	contentDisposition string
	contentEncoding    string
	contentLanguage    string
	metadata           map[string]string
	uploadHeaders      map[string]string
	data               []byte
}

var (
	ErrInvalidURL       = errors.New("invalid URL")
	ErrUnsupportedScheme = errors.New("only http and https schemes are supported")
	ErrInvalidPath       = errors.New("invalid file path")
)

// validateURL checks if a URL is valid and uses a safe scheme
func validateURL(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	
	// Only allow http and https schemes to prevent SSRF attacks
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrUnsupportedScheme
	}
	
	return nil
}

// validatePath checks if a file path is safe and doesn't contain directory traversal
func validatePath(path string) error {
	// Clean the path to resolve any .. or . components
	cleaned := filepath.Clean(path)
	
	// Check for directory traversal attempts
	if strings.Contains(cleaned, "..") {
		return ErrInvalidPath
	}
	
	return nil
}

func NewUpload(path string, storageClass proto.ObjectType) (*EntityObject, error) {
	if err := validatePath(path); err != nil {
		return nil, err
	}
	return &EntityObject{path: path, metadata: make(map[string]string), storageClass: storageClass}, nil
}

func NewUploadFromURL(path, remoteUrl string, storageClass proto.ObjectType) (*EntityObject, error) {
	if err := validatePath(path); err != nil {
		return nil, err
	}
	if err := validateURL(remoteUrl); err != nil {
		return nil, err
	}
	
	eo := &EntityObject{path: path, metadata: make(map[string]string), storageClass: storageClass}
	data, err := getRemoteFile(remoteUrl)
	if err != nil {
		return nil, err
	}
	fileContent, readErr := io.ReadAll(data)
	if readErr != nil {
		return nil, readErr
	}
	eo.SetData(fileContent)
	return eo, nil
}

func (e *EntityObject) SetPublic(public bool) {
	e.public = public
}

func (e *EntityObject) SetExpiry(expiry time.Time) {
	e.expiry = expiry
}

func (e *EntityObject) SetContentType(contentType string) {
	e.contentType = contentType
}

func (e *EntityObject) SetContentDisposition(contentDisposition string) {
	e.contentDisposition = contentDisposition
}

func (e *EntityObject) SetContentEncoding(contentEncoding string) {
	e.contentEncoding = contentEncoding
}

func (e *EntityObject) SetContentLanguage(contentLanguage string) {
	e.contentLanguage = contentLanguage
}

func (e *EntityObject) SetMeta(key, value string) {
	if e.metadata == nil {
		e.metadata = make(map[string]string)
	}
	e.metadata[key] = value
}

func (e *EntityObject) SetData(data []byte) { e.data = data }

func (e *EntityObject) GetPath() string {
	return e.path
}

func (e *EntityObject) GetUploadURL() string {
	return e.uploadURL
}

func (e *EntityObject) ReadyForUpload() bool {
	return e.uploadURL != ""
}

// secureHTTPClient returns an HTTP client with secure default settings
func secureHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: false,
		},
	}
}

func (e *EntityObject) Upload(content io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPut, e.GetUploadURL(), content)
	if err != nil {
		return nil, err
	}
	if e.uploadHeaders != nil {
		for k, v := range e.uploadHeaders {
			req.Header.Set(k, v)
		}
	}
	client := secureHTTPClient()
	return client.Do(req)
}

func (e *EntityObject) CopyFromURL(source string) (*http.Response, error) {
	if e.GetUploadURL() == "" {
		return nil, errors.New("upload URL is empty; call the API to initialize upload first")
	}
	
	if err := validateURL(source); err != nil {
		return nil, err
	}

	src, err := getRemoteFile(source)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, e.GetUploadURL(), src)
	if err != nil {
		return nil, err
	}

	// If we have headers provided for the upload, apply them
	if e.uploadHeaders != nil {
		for k, v := range e.uploadHeaders {
			req.Header.Set(k, v)
		}
	}

	client := secureHTTPClient()
	return client.Do(req)
}

func getRemoteFile(urlStr string) (io.Reader, error) {
	if err := validateURL(urlStr); err != nil {
		return nil, err
	}
	
	// Fetch the source content
	client := secureHTTPClient()
	srcResp, err := client.Get(urlStr)
	if err != nil {
		return nil, err
	}
	// Ensure the source response body is closed after we finish the upload request
	// (http.Client.Do below will read from this body before the function returns)
	defer srcResp.Body.Close()

	if srcResp.StatusCode < 200 || srcResp.StatusCode >= 300 {
		b, err := io.ReadAll(srcResp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to download source: status %s (could not read body: %v)", srcResp.Status, err)
		}
		return nil, fmt.Errorf("failed to download source: status %s body: %s", srcResp.Status, string(b))
	}

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, srcResp.Body)
	return buf, err
}

func (e *EntityObject) UploadToJson(content interface{}) (*http.Response, error) {
	jsn, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	return e.Upload(bytes.NewReader(jsn))
}

func UploadError(resp *http.Response, err error) error {
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		bdy, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return readErr
		}
		return errors.New("upload failed, status code: " + string(rune(resp.StatusCode)) + " body: " + string(bdy))
	}
	return nil
}

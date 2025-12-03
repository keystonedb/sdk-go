package keystone

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
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

func NewUpload(path string, storageClass proto.ObjectType) *EntityObject {
	return &EntityObject{path: path, metadata: make(map[string]string), storageClass: storageClass}
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
	return http.DefaultClient.Do(req)
}

func (e *EntityObject) CopyFromURL(source string) (*http.Response, error) {
	if e.GetUploadURL() == "" {
		return nil, errors.New("upload URL is empty; call the API to initialize upload first")
	}

	// Fetch the source content
	srcResp, err := http.Get(source)
	if err != nil {
		return nil, err
	}
	// Ensure the source response body is closed after we finish the upload request
	// (http.Client.Do below will read from this body before the function returns)
	defer srcResp.Body.Close()

	if srcResp.StatusCode < 200 || srcResp.StatusCode >= 300 {
		b, _ := io.ReadAll(srcResp.Body)
		return nil, errors.New("failed to download source: status " + srcResp.Status + " body: " + string(b))
	}

	// Prepare the upload PUT request with the downloaded body as content
	req, err := http.NewRequest(http.MethodPut, e.GetUploadURL(), srcResp.Body)
	if err != nil {
		return nil, err
	}

	// If we have headers provided for the upload, apply them
	if e.uploadHeaders != nil {
		for k, v := range e.uploadHeaders {
			req.Header.Set(k, v)
		}
	}

	// Preserve the content-type from the source if not explicitly set already
	if req.Header.Get("Content-Type") == "" {
		if ct := srcResp.Header.Get("Content-Type"); ct != "" {
			req.Header.Set("Content-Type", ct)
		}
	}

	// If the server provided a content-length, forward it (optional; http will chunk if absent)
	if cl := srcResp.Header.Get("Content-Length"); cl != "" {
		req.Header.Set("Content-Length", cl)
	}

	return http.DefaultClient.Do(req)
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

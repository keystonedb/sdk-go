package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"io"
	"net/http"
	"time"
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

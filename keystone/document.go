package keystone

import "github.com/keystonedb/sdk-go/proto"

type Document struct {
	RevisionID string
	Data       []byte
	Meta       map[string]string
	appendMeta map[string]string
	removeMeta []string
}

func (d *Document) apply(mutate *proto.MutateRequest) {
	mutate.Mutation.Document = &proto.EntityDocument{
		RevisionId: d.RevisionID,
		Meta:       d.Meta,
		Data:       d.Data,
		AppendMeta: d.appendMeta,
		RemoveMeta: d.removeMeta,
	}
}

func (d *Document) ObserveMutation(response *proto.MutateResponse) {
	if d.RevisionID == "" && response.DocumentRevisionId != "" {
		d.RevisionID = response.DocumentRevisionId
	}
}

func NewDocument(data []byte, meta map[string]string) *Document {
	return &Document{
		Data: data,
		Meta: meta,
	}
}

func UpdateDocument(revisionID string) *Document {
	return &Document{
		RevisionID: revisionID,
	}
}

func (d *Document) AppendMeta(key, value string) {
	if d.appendMeta == nil {
		d.appendMeta = make(map[string]string)
	}
	d.appendMeta[key] = value
}

func (d *Document) RemoveMeta(key string) {
	d.removeMeta = append(d.removeMeta, key)
}

func (d *Document) SetMeta(meta map[string]string) {
	d.Meta = meta
	d.removeMeta = nil
	d.appendMeta = nil
}

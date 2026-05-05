package keystone

import "github.com/keystonedb/sdk-go/proto"

// Link represents a URL with display text and target behaviour.
type Link struct {
	Href      string `json:"href"`
	Title     string `json:"title"`
	NewWindow bool   `json:"new_window"`
}

func NewLink(href, title string, newWindow bool) *Link {
	return &Link{
		Href:      href,
		Title:     title,
		NewWindow: newWindow,
	}
}

func (l *Link) String() string {
	if l == nil {
		return ""
	}
	if l.Title != "" {
		return l.Title
	}
	return l.Href
}

func (l *Link) IsZero() bool {
	return l == nil || (l.Href == "" && l.Title == "" && !l.NewWindow)
}

func (l *Link) MarshalValue() (*proto.Value, error) {
	if l.IsZero() {
		return nil, nil
	}
	return &proto.Value{
		Text: l.Title,
		Bool: l.NewWindow,
		Raw:  []byte(l.Href),
	}, nil
}

func (l *Link) UnmarshalValue(value *proto.Value) error {
	if value != nil {
		l.Href = string(value.GetRaw())
		l.Title = value.GetText()
		l.NewWindow = value.GetBool()
	}
	return nil
}

func (l *Link) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Mixed, ExtendedType: proto.Property_URL}
}

package keystone

import (
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

func TestLink_MarshalValue(t *testing.T) {
	link := NewLink("https://example.com/docs", "Docs", true)

	value, err := link.MarshalValue()
	if err != nil {
		t.Fatalf("MarshalValue returned error: %v", err)
	}
	if value.GetText() != "Docs" {
		t.Fatalf("Text = %q, want %q", value.GetText(), "Docs")
	}
	if !value.GetBool() {
		t.Fatal("Bool = false, want true")
	}
	if string(value.GetRaw()) != "https://example.com/docs" {
		t.Fatalf("Raw = %q, want %q", string(value.GetRaw()), "https://example.com/docs")
	}

	def := link.PropertyDefinition()
	if def.DataType != proto.Property_Mixed {
		t.Fatalf("DataType = %v, want %v", def.DataType, proto.Property_Mixed)
	}
	if def.ExtendedType != proto.Property_URL {
		t.Fatalf("ExtendedType = %v, want %v", def.ExtendedType, proto.Property_URL)
	}
}

func TestLink_UnmarshalValue(t *testing.T) {
	var link Link

	err := link.UnmarshalValue(&proto.Value{
		Text: "Docs",
		Bool: true,
		Raw:  []byte("https://example.com/docs"),
	})
	if err != nil {
		t.Fatalf("UnmarshalValue returned error: %v", err)
	}
	if link.Title != "Docs" {
		t.Fatalf("Title = %q, want %q", link.Title, "Docs")
	}
	if !link.NewWindow {
		t.Fatal("NewWindow = false, want true")
	}
	if link.Href != "https://example.com/docs" {
		t.Fatalf("Href = %q, want %q", link.Href, "https://example.com/docs")
	}
}

func TestMixed_LinkRoundTrip(t *testing.T) {
	link := Link{
		Href:      "https://example.com/docs",
		Title:     "Docs",
		NewWindow: true,
	}

	mixed := NewMixed(link)
	if mixed.String() != "Docs" {
		t.Fatalf("String = %q, want %q", mixed.String(), "Docs")
	}
	if !mixed.Bool() {
		t.Fatal("Bool = false, want true")
	}
	if string(mixed.Raw()) != "https://example.com/docs" {
		t.Fatalf("Raw = %q, want %q", string(mixed.Raw()), "https://example.com/docs")
	}

	decoded := mixed.Link()
	if decoded == nil {
		t.Fatal("Link() returned nil")
	}
	if *decoded != link {
		t.Fatalf("Link() = %+v, want %+v", *decoded, link)
	}
}

func TestAKV_Link(t *testing.T) {
	prop := AKV("docs", NewLink("https://example.com/docs", "Docs", true))

	if prop.Property == nil {
		t.Fatal("Property is nil")
	}
	if prop.Property.DataType != proto.Property_Mixed {
		t.Fatalf("DataType = %v, want %v", prop.Property.DataType, proto.Property_Mixed)
	}
	if prop.Property.ExtendedType != proto.Property_URL {
		t.Fatalf("ExtendedType = %v, want %v", prop.Property.ExtendedType, proto.Property_URL)
	}
	if prop.Value == nil {
		t.Fatal("Value is nil")
	}
	if prop.Value.GetText() != "Docs" {
		t.Fatalf("Text = %q, want %q", prop.Value.GetText(), "Docs")
	}
	if !prop.Value.GetBool() {
		t.Fatal("Bool = false, want true")
	}
	if string(prop.Value.GetRaw()) != "https://example.com/docs" {
		t.Fatalf("Raw = %q, want %q", string(prop.Value.GetRaw()), "https://example.com/docs")
	}
}

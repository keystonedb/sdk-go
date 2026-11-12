package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
	"slices"
	"testing"
)

func TestPIITagSetsOption(t *testing.T) {
	for _, tag := range []string{"pii", "personal", "gdpr"} {
		field := reflect.StructField{Name: "Profile", Tag: reflect.StructTag(`keystone:",` + tag + `"`)}
		_, tagged := ReflectProperty(field, "")
		if !slices.Contains(tagged.Options, proto.Property_PII) {
			t.Fatalf("%s did not set PII option", tag)
		}
		url := mergeDefinitions(tagged, proto.PropertyDefinition{DataType: proto.Property_SecureText, ExtendedType: proto.Property_URL})
		if url.ExtendedType != proto.Property_URL || url.DataType != proto.Property_SecureText || !slices.Contains(url.Options, proto.Property_PII) {
			t.Fatalf("URL classification lost: %+v", url)
		}
		if tagged.ExtendedType != proto.Property_Personal {
			t.Fatal("legacy personal classification changed")
		}
	}
	_, ordinary := ReflectProperty(reflect.StructField{Name: "Profile"}, "")
	if slices.Contains(ordinary.Options, proto.Property_PII) {
		t.Fatal("untagged field marked personal")
	}
}

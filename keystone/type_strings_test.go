package keystone

import (
	"cmp"
	"github.com/keystonedb/sdk-go/proto"
	"testing"
)

func Test_String(t *testing.T) {
	str := String("Hello, World!")
	pVal, err := str.MarshalValue()
	if err != nil {
		t.Errorf("Error marshalling value: %v", err)
	}

	if pVal.GetText() != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got %v", pVal.GetText())
	}

	var str2 String
	err = str2.UnmarshalValue(pVal)
	if err != nil {
		t.Errorf("Error unmarshalling value: %v", err)
	}

	if cmp.Compare(str2, str) != 0 {
		t.Errorf("Expected %v, got %v", str, str2)
	}

	if str2.PropertyDefinition().DataType != proto.Property_Text {
		t.Errorf("Expected proto.Property_Text, got %v", str2.PropertyDefinition().DataType)
	}
}

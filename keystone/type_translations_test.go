package keystone

import (
	"encoding/json"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

func Test_Translations(t *testing.T) {
	// Test basic Replace and Get
	translations := &Translations{}
	translations.Replace(map[string]*Translation{
		"en": NewTranslation("Hello"),
		"fr": NewTranslation("Bonjour"),
		"es": NewTranslation("Hola"),
	})

	if text, ok := translations.Get("en"); !ok || text.String() != "Hello" {
		t.Errorf("Expected 'Hello', got %v (ok: %v)", text.String(), ok)
	}

	if text, ok := translations.Get("fr"); !ok || text.String() != "Bonjour" {
		t.Errorf("Expected 'Bonjour', got %v (ok: %v)", text.String(), ok)
	}

	// Test Add
	translations.Add("de", "Hallo")
	if text, ok := translations.Get("de"); !ok || text.String() != "Hallo" {
		t.Errorf("Expected 'Hallo', got %v (ok: %v)", text.String(), ok)
	}

	// Test Remove
	translations.Remove("es")
	if _, ok := translations.Get("es"); ok {
		t.Errorf("Expected 'es' to be removed, but it still exists")
	}

	// Test All
	all := translations.All()
	expectedCount := 3 // en, fr, de (es removed)
	if len(all) != expectedCount {
		t.Errorf("Expected %d translations, got %d", expectedCount, len(all))
	}

	// Test MarshalValue
	pVal, err := translations.MarshalValue()
	if err != nil {
		t.Errorf("Error marshalling Value: %v", err)
	}

	if pVal.Array == nil || pVal.Array.GetKeyValue() == nil {
		t.Errorf("Expected KeyValue to be set in Array")
	}

	// Test UnmarshalValue
	var translations2 Translations
	err = translations2.UnmarshalValue(pVal)
	if err != nil {
		t.Errorf("Error unmarshalling Value: %v", err)
	}

	all2 := translations2.All()
	if len(all2) != expectedCount {
		t.Errorf("After unmarshal, expected %d translations, got %d", expectedCount, len(all2))
	}

	// Verify individual values
	if text, ok := translations2.Get("en"); !ok || text.String() != "Hello" {
		t.Errorf("After unmarshal, expected 'Hello' for 'en', got %v (ok: %v)", text.String(), ok)
	}

	// Test PropertyDefinition
	if translations2.PropertyDefinition().DataType != proto.Property_KeyValue {
		t.Errorf("Expected proto.Property_KeyValue, got %v", translations2.PropertyDefinition().DataType)
	}

	// Test IsZero
	if translations2.IsZero() {
		t.Errorf("Expected translations2 to not be zero")
	}

	empty := &Translations{}
	if !empty.IsZero() {
		t.Errorf("Expected empty translations to be zero")
	}
}

func Test_Translations_AddRemove(t *testing.T) {
	translations := &Translations{}

	// Add some translations
	translations.Add("en", "Hello")
	translations.Add("fr", "Bonjour")

	all := translations.All()
	if len(all) != 2 {
		t.Errorf("Expected 2 translations, got %d", len(all))
	}

	// Remove one
	translations.Remove("en")
	all = translations.All()
	if len(all) != 1 {
		t.Errorf("Expected 1 translation after remove, got %d", len(all))
	}

	if _, ok := all["en"]; ok {
		t.Errorf("Expected 'en' to be removed")
	}

	if _, ok := all["fr"]; !ok {
		t.Errorf("Expected 'fr' to still exist")
	}
}

func Test_Translations_Update(t *testing.T) {
	translations := &Translations{}
	translations.Replace(map[string]*Translation{"en": NewTranslation("Hello")})

	// Update existing translation
	translations.Add("en", "Hi")

	text, ok := translations.Get("en")
	if !ok || text.String() != "Hi" {
		t.Errorf("Expected 'Hi' after update, got %v (ok: %v)", text, ok)
	}
}

func Test_Translations_MarshalUnmarshal(t *testing.T) {
	translations := &Translations{}
	translations.Replace(map[string]*Translation{
		"en": NewTranslation("Hello"),
		"fr": NewTranslation("Bonjour"),
	})
	translations.Add("de", "Hallo")
	translations.Remove("fr")

	// Marshal
	pVal, err := translations.MarshalValue()
	if err != nil {
		t.Fatalf("Error marshalling: %v", err)
	}

	// Unmarshal
	var translations2 Translations
	err = translations2.UnmarshalValue(pVal)
	if err != nil {
		t.Fatalf("Error unmarshalling: %v", err)
	}

	// Verify
	all := translations2.All()
	if len(all) != 2 {
		t.Errorf("Expected 2 translations (en, de), got %d: %v", len(all), all)
	}

	if text, ok := translations2.Get("en"); !ok || text.String() != "Hello" {
		t.Errorf("Expected 'Hello' for 'en', got %v", text.String())
	}

	if text, ok := translations2.Get("de"); !ok || text.String() != "Hallo" {
		t.Errorf("Expected 'Hallo' for 'de', got %v", text.String())
	}

	if _, ok := translations2.Get("fr"); ok {
		t.Errorf("Expected 'fr' to be removed")
	}
}

func Test_Translations_JSON(t *testing.T) {
	translations := &Translations{}
	translations.Add("en", "Hello")
	translations.Add("fr", "Bonjour")
	translations.Add("de", "Hallo")
	translations.Remove("fr")

	jsnVal, err := json.Marshal(translations)
	if err != nil {
		t.Fatalf("Error marshalling: %v", err)
	}

	if string(jsnVal) != "{\"de\":{\"s\":\"Hallo\"},\"en\":{\"s\":\"Hello\"}}" {
		t.Errorf("Expected specific json string, got %v", string(jsnVal))
	}

	var translations2 Translations
	err = json.Unmarshal(jsnVal, &translations2)
	if err != nil {
		t.Fatalf("Error unmarshalling: %v", err)
	}
	all := translations2.All()
	if len(all) != 2 {
		t.Errorf("Expected 2 translations (en, de), got %d: %v", len(all), all)
	}
	if text, ok := translations2.Get("en"); !ok || text.String() != "Hello" {
		t.Errorf("Expected 'Hello' for 'en', got %v", text.String())
	}
	if text, ok := translations2.Get("de"); !ok || text.String() != "Hallo" {
		t.Errorf("Expected 'Hallo' for 'de', got %v", text.String())
	}
	if _, ok := translations2.Get("fr"); ok {
		t.Errorf("Expected 'fr' to be removed")
	}
}
